package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/blcvn/backend/services/ai-proxy-service/cache"
	"github.com/blcvn/backend/services/ai-proxy-service/providers"
	"github.com/blcvn/backend/services/ai-proxy-service/resilience"
	"github.com/blcvn/backend/services/ai-proxy-service/router"
	aimodel "github.com/blcvn/kratos-proto/go/ai-model"
)

// ProxyUsecase implements the core business logic for AI Proxy
type ProxyUsecase struct {
	router         *router.Router
	cache          *cache.RedisCache
	circuitBreaker *resilience.CircuitBreakerManager
	modelService   aimodel.AIModelServiceClient
}

// NewProxyUsecase creates a new proxy usecase
func NewProxyUsecase(
	router *router.Router,
	cache *cache.RedisCache,
	circuitBreaker *resilience.CircuitBreakerManager,
	modelService aimodel.AIModelServiceClient,
) *ProxyUsecase {
	return &ProxyUsecase{
		router:         router,
		cache:          cache,
		circuitBreaker: circuitBreaker,
		modelService:   modelService,
	}
}

// Complete handles a completion request with caching, routing, and circuit breaker
func (u *ProxyUsecase) Complete(ctx context.Context, modelID string, req *providers.CompletionRequest) (*providers.CompletionResponse, bool, error) {
	startTime := time.Now()

	// 1. Check cache
	if u.cache.ShouldCache(req) {
		cacheKey := u.cache.GenerateKey(modelID, req)
		cached, err := u.cache.Get(ctx, cacheKey)
		if err == nil && cached != nil {
			// Cache hit
			return cached, true, nil
		}
	}

	// 2. Route to provider
	provider, _, err := u.router.Route(ctx, modelID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to route request: %w", err)
	}

	// 3. Execute with circuit breaker
	result, err := u.circuitBreaker.Execute(modelID, func() (interface{}, error) {
		// Set API key dynamically (hacky but works for now)
		// In production, we'd refactor provider interface to accept apiKey in Complete()
		return provider.Complete(ctx, req)
	})

	if err != nil {
		return nil, false, fmt.Errorf("completion failed: %w", err)
	}

	response := result.(*providers.CompletionResponse)

	// 4. Cache response (if deterministic)
	if u.cache.ShouldCache(req) {
		cacheKey := u.cache.GenerateKey(modelID, req)
		if err := u.cache.Set(ctx, cacheKey, response, 1*time.Hour); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to cache response: %v\n", err)
		}
	}

	// 5. Log usage to AI Model Service
	latencyMs := int32(time.Since(startTime).Milliseconds())
	_, err = u.modelService.LogUsage(ctx, &aimodel.LogUsageRequest{
		Payload: &aimodel.LogUsagePayload{
			ModelId:    modelID,
			TokensUsed: int64(response.TokensUsed),
			LatencyMs:  latencyMs,
			Status:     aimodel.UsageStatus_SUCCESS_STATUS,
		},
	})
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to log usage: %v\n", err)
	}

	return response, false, nil
}

// HealthCheck checks the health of all providers
func (u *ProxyUsecase) HealthCheck(ctx context.Context) (bool, error) {
	// For now, just return healthy
	// In production, we'd check circuit breaker states and provider health
	return true, nil
}
