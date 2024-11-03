package main

import (
	"authorization_flow_oauth/config"
	"authorization_flow_oauth/internal/handler"
	"authorization_flow_oauth/internal/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

func main() {
	config := config.Config{
		KeycloakURL: "http://localhost:8080",          // Adjust this to your Keycloak server address
		Realm:           "dev",                            // Your Keycloak realm name
		ClientID:        "pos",                            // Your client ID in Keycloak
		RedirectURL:     "http://localhost:8081/callback", // This should match a valid redirect URI in your Keycloak client settings
		ClientSecret:    "YlLT3yNV7EyiTPYLuxcXs1fAiExKHNFx",
	}
	// context
	ctx := context.Background()
	// In a production environment, use a secure key management system
	// This is just for demonstration purposes
	store := sessions.NewCookieStore([]byte("secret-key-replace-this-in-production"))
	oidcClient, err  := utils.NewOIDCClient(ctx, &config)
	if err != nil {
        log.Fatalf("Failed to create OIDC client: %v", err)
	}
	http.HandleFunc("/login-keycloak", oidcClient.HandleLogin)
	http.HandleFunc("/", handler.HomeHandler)
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		handler.CallbackHandler(ctx, &config, w, r, store, oidcClient)
	})
	http.HandleFunc("/dashboard",func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "auth-session")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Check if user is authenticated
    email, ok := session.Values["user_email"].(string)
    if !ok {
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    // Get additional user info if stored in session
    name, _ := session.Values["user_name"].(string)

    w.WriteHeader(http.StatusOK)
    w.Header().Set("Content-Type", "application/json")
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "email": email,
        "name": name,
    })
	})

	fmt.Println("Server is starting on port 8081...")
	err = http.ListenAndServe(":8081", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
