package usecases

import (
	"context"
	"fmt"

	"github.com/blcvn/backend/services/ai-proxy-service/common/errors"
	"github.com/blcvn/backend/services/ai-proxy-service/entities"
	model_pb "github.com/blcvn/kratos-proto/go/ai-model"
)

type iAIModelClient interface {
	GetCredentials(ctx context.Context, modelID string) (*model_pb.Credentials, error)
	GetModel(ctx context.Context, modelID string) (*model_pb.AIModel, error)
	CheckQuota(ctx context.Context, modelID string, tokens int32) (bool, error)
	LogUsage(ctx context.Context, modelID string, promptTokens, completionTokens int32) error
}

// ProxyUsecase implements the core business logic for AI Proxy
type ProxyUsecase struct {
	modelClient iAIModelClient
	providers   map[string]entities.LLMProvider
}

func NewProxyUsecase(modelClient iAIModelClient) *ProxyUsecase {
	return &ProxyUsecase{
		modelClient: modelClient,
		providers:   make(map[string]entities.LLMProvider),
	}
}

func (u *ProxyUsecase) RegisterProvider(name string, provider entities.LLMProvider) {
	u.providers[name] = provider
}

func (u *ProxyUsecase) Complete(ctx context.Context, req *entities.CompletionRequest) (*entities.CompletionResponse, errors.BaseError) {
	// 1. Check Quota
	// tạm thời bỏ logic check quota
	// allowed, err := u.modelClient.CheckQuota(ctx, req.ModelID, req.MaxTokens)
	// if err != nil {
	// 	return nil, errors.Internal(err)
	// }
	// if !allowed {
	// 	return nil, errors.RateLimit("quota exceeded for this model")
	// }

	// 2. Get Model Info (to know provider)
	model, err := u.modelClient.GetModel(ctx, req.ModelID)
	if err != nil {
		return nil, errors.Internal(err)
	}

	// 3. Get Credentials
	creds, err := u.modelClient.GetCredentials(ctx, req.ModelID)
	if err != nil {
		return nil, errors.Internal(err)
	}

	// Inject credentials into request
	req.APIKey = creds.ApiKey
	req.BaseURL = creds.BaseUrl

	// 4. Get Provider
	provider, ok := u.providers[model.Provider]
	if !ok {
		return nil, errors.BadRequest(fmt.Sprintf("unsupported provider: %s", model.Provider))
	}

	// 5. Call LLM
	// TODO: Inject creds into provider before calling, or pass creds to Complete
	resp, err := provider.Complete(ctx, req)
	if err != nil {
		return nil, errors.Internal(err)
	}

	// 5. Log Usage
	_ = u.modelClient.LogUsage(ctx, req.ModelID, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)

	return resp, nil
}

func (u *ProxyUsecase) HealthCheck(ctx context.Context) (bool, error) {
	// Simple health check: verify DB/Redis connectivity if possible.
	// For now, return true as basic check.
	// TODO: Add proper health checks
	return true, nil
}

func (u *ProxyUsecase) StreamComplete(ctx context.Context, req *entities.CompletionRequest, callback func(*entities.StreamResponse) error) errors.BaseError {
	// 1. Check Quota
	// allowed, err := u.modelClient.CheckQuota(ctx, req.ModelID, req.MaxTokens)
	// if err != nil {
	// 	return errors.Internal(err)
	// }
	// if !allowed {
	// 	return errors.RateLimit("quota exceeded for this model")
	// }

	// 2. Get Model Info
	model, err := u.modelClient.GetModel(ctx, req.ModelID)
	if err != nil {
		return errors.Internal(err)
	}

	// 3. Get Credentials
	creds, err := u.modelClient.GetCredentials(ctx, req.ModelID)
	if err != nil {
		return errors.Internal(err)
	}

	// Inject credentials into request
	req.APIKey = creds.ApiKey
	req.BaseURL = creds.BaseUrl

	// 4. Get Provider
	provider, ok := u.providers[model.Provider]
	if !ok {
		return errors.BadRequest(fmt.Sprintf("unsupported provider: %s", model.Provider))
	}

	// 4. Stream LLM
	var totalPrompt, totalCompletion int32
	err = provider.StreamComplete(ctx, req, func(sr *entities.StreamResponse) error {
		if sr.Usage != nil {
			totalPrompt = sr.Usage.PromptTokens
			totalCompletion = sr.Usage.CompletionTokens
		}
		return callback(sr)
	})

	if err != nil {
		return errors.Internal(err)
	}

	// 5. Log Usage
	if totalPrompt > 0 || totalCompletion > 0 {
		_ = u.modelClient.LogUsage(ctx, req.ModelID, totalPrompt, totalCompletion)
	}

	return nil
}
