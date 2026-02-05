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
	CheckQuota(ctx context.Context, modelID string, tokens int32) (bool, error)
	LogUsage(ctx context.Context, modelID string, promptTokens, completionTokens int32) error
}

type aiProxyUsecase struct {
	modelClient iAIModelClient
	providers   map[string]entities.LLMProvider
}

func NewAIProxyUsecase(modelClient iAIModelClient) *aiProxyUsecase {
	return &aiProxyUsecase{
		modelClient: modelClient,
		providers:   make(map[string]entities.LLMProvider),
	}
}

func (u *aiProxyUsecase) RegisterProvider(name string, provider entities.LLMProvider) {
	u.providers[name] = provider
}

func (u *aiProxyUsecase) Complete(ctx context.Context, req *entities.CompletionRequest) (*entities.CompletionResponse, errors.BaseError) {
	// 1. Check Quota
	allowed, err := u.modelClient.CheckQuota(ctx, req.ModelID, req.MaxTokens)
	if err != nil {
		return nil, errors.Internal(err)
	}
	if !allowed {
		return nil, errors.RateLimit("quota exceeded for this model")
	}

	// 2. Get Credentials
	creds, err := u.modelClient.GetCredentials(ctx, req.ModelID)
	if err != nil {
		return nil, errors.Internal(err)
	}

	// 3. Get Provider
	provider, ok := u.providers[creds.Provider]
	if !ok {
		return nil, errors.BadRequest(fmt.Sprintf("unsupported provider: %s", creds.Provider))
	}

	// 4. Call LLM
	// TODO: Inject creds into provider before calling, or pass creds to Complete
	resp, err := provider.Complete(ctx, req)
	if err != nil {
		return nil, errors.Internal(err)
	}

	// 5. Log Usage
	_ = u.modelClient.LogUsage(ctx, req.ModelID, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)

	return resp, nil
}

func (u *aiProxyUsecase) StreamComplete(ctx context.Context, req *entities.CompletionRequest, callback func(*entities.StreamResponse) error) errors.BaseError {
	// 1. Check Quota
	allowed, err := u.modelClient.CheckQuota(ctx, req.ModelID, req.MaxTokens)
	if err != nil {
		return errors.Internal(err)
	}
	if !allowed {
		return errors.RateLimit("quota exceeded for this model")
	}

	// 2. Get Credentials
	creds, err := u.modelClient.GetCredentials(ctx, req.ModelID)
	if err != nil {
		return errors.Internal(err)
	}

	// 3. Get Provider
	provider, ok := u.providers[creds.Provider]
	if !ok {
		return errors.BadRequest(fmt.Sprintf("unsupported provider: %s", creds.Provider))
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
