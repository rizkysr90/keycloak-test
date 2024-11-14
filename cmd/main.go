package main

import (
	"context"
	"fmt"
	"log"

	"authorization_flow_oauth/internal/config"
	"authorization_flow_oauth/internal/handler/authhandler"
	rds "authorization_flow_oauth/internal/store/redis"
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
	serverAddr := fmt.Sprintf("%s:%s", cfg.AppHost, cfg.AppPort)
	// context
	ctx := context.Background()

	authOptions := []auth.Option{
		auth.WithClientSecret(cfg.Auth.ClientSecret),
		auth.WithRealmKeycloak(cfg.Auth.Realm),
	}
	authClient, err := auth.New(
		ctx,
		cfg.Auth.BaseURL,
		cfg.Auth.ClientID,
		cfg.Auth.RedirectURL,
		authOptions...,
	)
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
	// Load HTML templates from internal/templates
	// Using relative path from where you run the application
	r.LoadHTMLGlob("../internal/templates/*/*.tmpl")

	// Auth handler
	authStore := rds.NewAuthRedisManager(&rds.Config{RedisClient: redisClient})
	authHandler := authhandler.New(cfg, serverAddr, authClient, authStore)
	r.GET("/login", authHandler.RenderLoginPage)
	r.GET("/login-keycloak", authHandler.RedirectToKeycloak)
	r.GET("/callback-auth", authHandler.Callback)

	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
