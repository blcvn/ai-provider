package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the AI Proxy Service
type Config struct {
	// Server
	Port string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// AI Model Service
	AIModelServiceAddr string

	// Circuit Breaker
	CircuitBreakerMaxRequests uint32
	CircuitBreakerInterval    int // seconds
	CircuitBreakerTimeout     int // seconds

	// Cache
	CacheTTL int // seconds

	// Metrics
	MetricsPort string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		Port:                      getEnv("PORT", "8087"),
		RedisAddr:                 getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:             getEnv("REDIS_PASSWORD", ""),
		RedisDB:                   getEnvAsInt("REDIS_DB", 0),
		AIModelServiceAddr:        getEnv("AI_MODEL_SERVICE_ADDR", "localhost:8085"),
		CircuitBreakerMaxRequests: uint32(getEnvAsInt("CIRCUIT_BREAKER_MAX_REQUESTS", 5)),
		CircuitBreakerInterval:    getEnvAsInt("CIRCUIT_BREAKER_INTERVAL", 60),
		CircuitBreakerTimeout:     getEnvAsInt("CIRCUIT_BREAKER_TIMEOUT", 60),
		CacheTTL:                  getEnvAsInt("CACHE_TTL", 3600),
		MetricsPort:               getEnv("METRICS_PORT", "9090"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
