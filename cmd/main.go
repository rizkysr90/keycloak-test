package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"authorization_flow_oauth/internal/config"
	"authorization_flow_oauth/internal/handler/login"
	"authorization_flow_oauth/pkg/auth"

	"github.com/gin-gonic/gin"
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
	authOptions := []auth.Option{
		auth.WithClientSecret(cfg.Auth.ClientSecret),
		auth.WithRealmKeycloak(cfg.Auth.Realm),
	}
	log.Println("HEREEE : ", cfg.Auth)
	authClient, err := auth.New(
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
	r := gin.Default()
	// Middleware to inject the base context
	r.Use(func(c *gin.Context) {
		// Attach the base context to the request
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	// Load HTML templates from internal/templates
	// Using relative path from where you run the application
	r.LoadHTMLGlob("../internal/templates/*/*.tmpl")
	r.GET("/login", login.Index)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/login-keycloak", func(c *gin.Context) {
		// login.WithKeycloak(c, authClient)
		c.Redirect(http.StatusFound, authClient.Oauth.AuthCodeURL("sadasad"))

	})
	// Modify the Run call to be more explicit
	serverAddr := fmt.Sprintf("%s:%s", "0.0.0.0", cfg.Port)
	log.Printf("Server starting on %s", serverAddr)
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
