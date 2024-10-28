package handler

import (
	"authorization_flow_oauth/config"
	"authorization_flow_oauth/internal/utils"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/sessions"
)

func CallbackHandler(config *config.Config, w http.ResponseWriter, r *http.Request, store *sessions.CookieStore) {
	session, err := store.Get(r, "auth-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Retrieve state from session
	sessionState, ok := session.Values["state"].(string)
	if !ok {
		http.Error(w, "Invalid session state", http.StatusBadRequest)
		return
	}

	// Retrieve state from URL parameters
	urlState := r.URL.Query().Get("state")
	// Compare states
	if sessionState != urlState {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Clear the used state from the session
	delete(session.Values, "state")
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// This is where you would make a request to Keycloak to exchange the authorization code for tokens
	// Extract the authorization code from the query parameters
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code found", http.StatusBadRequest)
		return
	}
	// Construct the token endpoint URL
	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token",
		config.KeycloakBaseURL, config.Realm)
	// Prepare the request body
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret) // Make sure to add ClientSecret to your Config struct
	data.Set("code", code)
	data.Set("redirect_uri", config.RedirectURI)
	// Make the request to exchange the code for tokens
	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

    if resp.StatusCode != http.StatusOK {
        log.Printf("Token exchange failed. Status: %d, Body: %s", resp.StatusCode, string(body))
        http.Error(w, fmt.Sprintf("Token exchange failed: %s", resp.Status), resp.StatusCode)
        return
    }

    // Parse the JSON response
    var result map[string]interface{}
    if err := json.Unmarshal(body, &result); err != nil {
        log.Printf("Error parsing token response: %v", err)
        http.Error(w, "Failed to parse token response", http.StatusInternalServerError)
        return
    }
	// / Construct the JWKS URL for Keycloak
	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs",
		config.KeycloakBaseURL, config.Realm)

	// Create key manager
	keyManager := utils.NewKeyManager(jwksURL)

	 // Extract kid from token
	 kid, err := ExtractKid(result["access_token"].(string))
	 if err != nil {
		 log.Fatalf("Failed to extract kid: %v", err)
	 }

	
	// Get public key for this kid
	publicKey, err := keyManager.GetPublicKey(kid)
	if err != nil {
		// Handle error getting public key
		panic(err.Error())
	}
	log.Println("HOREEE", *publicKey)
	// Now you can use this public key to validate your token
	claims := &utils.CustomClaims{}
	claims, err = utils.ValidateWithPublicKey(result["access_token"].(string), publicKey)
	if err != nil {
		fmt.Printf("Error validating token: %v\n", err)
		return
	}
	log.Println(claims)
	// Fetch user info using the access token
	userInfo, err := fetchUserInfo(config, result["access_token"].(string))
	if err != nil {
		log.Printf("Error fetching user info: %v", err)
		http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
		return
	}
	log.Println("Waduh : ", userInfo)
	
    // TODO: Store tokens securely (e.g., in an encrypted session)

    log.Printf("Token exchange successful. Access token: %s...", result["access_token"].(string))
    fmt.Fprintf(w, "Authorization successful! Tokens received.")
    
}

func fetchUserInfo(config *config.Config, accessToken string) ( map[string]interface{}, error) {
    userInfoURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/userinfo",
        config.KeycloakBaseURL, config.Realm)

    req, err := http.NewRequest("GET", userInfoURL, nil)
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", "Bearer "+accessToken)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("failed to fetch user info: %s", resp.Status)
    }
    var result map[string]interface{}

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return result, nil
}

// TokenHeader represents the JWT header structure
type TokenHeader struct {
	Kid string `json:"kid"`
	Alg string `json:"alg"`
}
// ExtractKid safely extracts the 'kid' from a JWT token header
func ExtractKid(tokenString string) (string, error) {
	// Split the token into parts
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token format")
	}

	// Decode the header part
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("failed to decode token header: %w", err)
	}

	// Parse the header
	var header TokenHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return "", fmt.Errorf("failed to parse token header: %w", err)
	}

	if header.Kid == "" {
		return "", fmt.Errorf("no 'kid' found in token header")
	}

	return header.Kid, nil
}