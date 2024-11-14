package rds

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config contains configuration for RedisAuthManager
type Config struct {
	RedisClient *redis.Client
	PrefixState string
	DefaultTTL  time.Duration
}
type RedisAuthManager struct {
	client      *redis.Client
	PrefixState string
	defaultTTL  time.Duration
}

func NewAuthRedisManager(config *Config) *RedisAuthManager {
	if config.PrefixState == "" {
		config.PrefixState = "st"
	}
	if config.DefaultTTL == 0 {
		config.DefaultTTL = 5 * time.Minute
	}
	return &RedisAuthManager{
		client:      config.RedisClient,
		PrefixState: config.PrefixState,
		defaultTTL:  config.DefaultTTL,
	}
}

func (r *RedisAuthManager) buildKeyState(state string) string {
	return fmt.Sprintf("%s:%s", r.PrefixState, state)
}

func (r *RedisAuthManager) SetState(ctx context.Context, state string) error {
	key := r.buildKeyState(state)
	expiration := r.defaultTTL

	err := r.client.Set(ctx, key, state, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set session in Redis: %w", err)
	}
	return nil
}

func (r *RedisAuthManager) GetState(
	ctx context.Context,
	state string,
) (string, error) {
	key := r.buildKeyState(state)
	stateData, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get state data from Redis: %w", err)
	}
	return stateData, nil
}
func (r *RedisAuthManager) DeleteState(
	ctx context.Context,
	state string,
) error {
	key := r.buildKeyState(state)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to remove state data from Redis: %w", err)
	}
	return nil
}
