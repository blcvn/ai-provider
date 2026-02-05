package router

import (
	"context"
	"fmt"
	"sync"

	"github.com/blcvn/backend/services/ai-proxy-service/providers"
	"github.com/blcvn/backend/services/ai-proxy-service/providers/anthropic"
	"github.com/blcvn/backend/services/ai-proxy-service/providers/local"
	"github.com/blcvn/backend/services/ai-proxy-service/providers/openai"
	aimodel "github.com/blcvn/kratos-proto/go/ai-model"
)

// LoadBalancer implements round-robin load balancing for API keys
type LoadBalancer struct {
	keys    []string
	current int
	mu      sync.Mutex
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(keys []string) *LoadBalancer {
	return &LoadBalancer{
		keys: keys,
	}
}

// NextKey returns the next API key in round-robin fashion
func (lb *LoadBalancer) NextKey() string {
	if len(lb.keys) == 0 {
		return ""
	}

	lb.mu.Lock()
	defer lb.mu.Unlock()

	key := lb.keys[lb.current]
	lb.current = (lb.current + 1) % len(lb.keys)
	return key
}

// Router handles routing requests to appropriate providers
type Router struct {
	modelService  aimodel.AIModelServiceClient
	providers     map[string]providers.LLMProvider
	loadBalancers map[string]*LoadBalancer
	mu            sync.RWMutex
}

// NewRouter creates a new router
func NewRouter(modelService aimodel.AIModelServiceClient) *Router {
	return &Router{
		modelService:  modelService,
		providers:     make(map[string]providers.LLMProvider),
		loadBalancers: make(map[string]*LoadBalancer),
	}
}

// Route selects the appropriate provider and API key for a model
func (r *Router) Route(ctx context.Context, modelID string) (providers.LLMProvider, string, error) {
	// Get model configuration from AI Model Service
	resp, err := r.modelService.GetModel(ctx, &aimodel.GetModelRequest{
		Id: modelID,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get model config: %w", err)
	}

	if resp.Result.Code != aimodel.ResultCode_SUCCESS {
		return nil, "", fmt.Errorf("failed to get model: %s", resp.Result.Message)
	}

	model := resp.Model
	provider := model.Provider

	// Get or create provider adapter
	providerAdapter, err := r.getOrCreateProvider(provider, model.ModelId, model.BaseUrl)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get provider: %w", err)
	}

	// Get API key (with load balancing if multiple keys exist)
	apiKey, err := r.getAPIKey(ctx, modelID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get API key: %w", err)
	}

	return providerAdapter, apiKey, nil
}

// getOrCreateProvider returns or creates a provider adapter
func (r *Router) getOrCreateProvider(providerType, modelID, baseURL string) (providers.LLMProvider, error) {
	r.mu.RLock()
	provider, exists := r.providers[modelID]
	r.mu.RUnlock()

	if exists {
		return provider, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if provider, exists := r.providers[modelID]; exists {
		return provider, nil
	}

	// Get API key for provider instantiation
	// Note: This is a temporary placeholder - API key will be passed from Route()
	apiKey := ""

	// Create new provider based on type
	var newProvider providers.LLMProvider
	var err error

	switch providerType {
	case "anthropic":
		newProvider, err = anthropic.NewClaudeProvider(apiKey, modelID)
	case "openai":
		newProvider, err = openai.NewGPTProvider(apiKey, modelID)
	case "ollama", "local":
		// For Ollama, use baseURL
		newProvider, err = local.NewOllamaProvider(baseURL, modelID)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	r.providers[modelID] = newProvider
	return newProvider, nil
}

// getAPIKey retrieves the API key for a model (with load balancing)
func (r *Router) getAPIKey(ctx context.Context, modelID string) (string, error) {
	// Get credentials from AI Model Service
	resp, err := r.modelService.GetCredentials(ctx, &aimodel.GetCredentialsRequest{
		ModelId: modelID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get credentials: %w", err)
	}

	if resp.Result.Code != aimodel.ResultCode_SUCCESS {
		return "", fmt.Errorf("failed to get credentials: %s", resp.Result.Message)
	}

	if resp.Credentials == nil {
		return "", fmt.Errorf("credentials not found")
	}

	return resp.Credentials.ApiKey, nil
}
