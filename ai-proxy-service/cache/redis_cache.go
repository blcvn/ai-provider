package cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/blcvn/backend/services/ai-proxy-service/providers"
	"github.com/go-redis/redis/v8"
)

// RedisCache implements caching for LLM completions
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(addr, password string, db int) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisCache{
		client: client,
	}
}

// GenerateKey generates a cache key from a completion request
func (c *RedisCache) GenerateKey(modelID string, req *providers.CompletionRequest) string {
	// Create a deterministic key from request parameters
	data := fmt.Sprintf("%s:%s:%f:%d:%f",
		modelID,
		req.Prompt,
		req.Temperature,
		req.MaxTokens,
		req.TopP,
	)

	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("cache:completion:%x", hash)
}

// Get retrieves a cached completion response
func (c *RedisCache) Get(ctx context.Context, key string) (*providers.CompletionResponse, error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var response providers.CompletionResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached response: %w", err)
	}

	return &response, nil
}

// Set stores a completion response in cache
func (c *RedisCache) Set(ctx context.Context, key string, response *providers.CompletionResponse, ttl time.Duration) error {
	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// ShouldCache determines if a request should be cached based on temperature
func (c *RedisCache) ShouldCache(req *providers.CompletionRequest) bool {
	// Only cache deterministic requests (temperature = 0)
	return req.Temperature == 0
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}
