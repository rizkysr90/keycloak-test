package main

import (
	"authorization_flow_oauth/internal/config"
	"authorization_flow_oauth/pkg/authclient"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
)

func main() {

	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("failed to load and parse config : %v", err)
		return
	}

	// context
	ctx := context.Background()

	// OAUTH
	authOptions := []authclient.Option{
		authclient.WithClientSecret(cfg.Auth.ClientSecret),
		authclient.WithRealmKeycloak(cfg.Auth.Realm),
	}
	authClient, err := authclient.New(
		ctx,
		cfg.Auth.BaseURL,
		cfg.Auth.ClientID,
		cfg.Auth.RedirectURL,
		authOptions...,
	)
	log.Println(authClient)
	if err != nil {
		log.Fatalf("Failed to initialize auth client : %v", err)
		return
	}
	// Redis
	redisClient := redis.NewClient(&cfg.RedisConfig)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
		return
	}
	defer redisClient.Close()

	// http.HandleFunc("/login-keycloak", oidcClient.HandleLogin)
	// http.HandleFunc("/", handler.HomeHandler)
	// http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
	// 	handler.CallbackHandler(ctx, &config, w, r, store, oidcClient)
	// })
	// http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
	// 	session, err := store.Get(r, "auth-session")
	// 	if err != nil {
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}

	// 	// Check if user is authenticated
	// 	email, ok := session.Values["user_email"].(string)
	// 	if !ok {
	// 		http.Redirect(w, r, "/login", http.StatusFound)
	// 		return
	// 	}

	// 	// Get additional user info if stored in session
	// 	name, _ := session.Values["user_name"].(string)

	// 	w.WriteHeader(http.StatusOK)
	// 	w.Header().Set("Content-Type", "application/json")

	// 	json.NewEncoder(w).Encode(map[string]interface{}{
	// 		"status": "success",
	// 		"email":  email,
	// 		"name":   name,
	// 	})
	// })

	fmt.Println("Server is starting on port 8081...")
	err = http.ListenAndServe(":8081", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
