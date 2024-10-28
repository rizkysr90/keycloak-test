package main

import (
	"authorization_flow_oauth/config"
	"authorization_flow_oauth/internal/handler"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

func main() {
	config := config.Config{
		KeycloakBaseURL: "http://localhost:8080",          // Adjust this to your Keycloak server address
		Realm:           "dev",                            // Your Keycloak realm name
		ClientID:        "pos",                            // Your client ID in Keycloak
		RedirectURI:     "http://localhost:8081/callback", // This should match a valid redirect URI in your Keycloak client settings
		ClientSecret:    "YlLT3yNV7EyiTPYLuxcXs1fAiExKHNFx",
	}
	// In a production environment, use a secure key management system
	// This is just for demonstration purposes
	store := sessions.NewCookieStore([]byte("secret-key-replace-this-in-production"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler.HomeHandler(&config, w, r, store)
	})
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		handler.CallbackHandler(&config, w, r, store)
	})

	fmt.Println("Server is starting on port 8081...")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
