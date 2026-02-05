package controllers

import (
	"context"

	"github.com/blcvn/backend/services/ai-proxy-service/providers"
	"github.com/blcvn/backend/services/ai-proxy-service/usecases"
	aiproxy "github.com/blcvn/kratos-proto/go/ai-proxy"
)

// ProxyController implements the AIProxyService gRPC interface
type ProxyController struct {
	aiproxy.UnimplementedAIProxyServiceServer
	usecase *usecases.ProxyUsecase
}

// NewProxyController creates a new proxy controller
func NewProxyController(usecase *usecases.ProxyUsecase) *ProxyController {
	return &ProxyController{
		usecase: usecase,
	}
}

// Complete handles completion requests
func (c *ProxyController) Complete(ctx context.Context, req *aiproxy.CompleteRequest) (*aiproxy.CompleteResponse, error) {
	if req.Payload == nil {
		return &aiproxy.CompleteResponse{
			Result: &aiproxy.Result{
				Code:    aiproxy.ResultCode_BAD_REQUEST,
				Message: "payload is required",
			},
		}, nil
	}

	// Convert proto request to provider request
	providerReq := &providers.CompletionRequest{
		Prompt:        req.Payload.Prompt,
		Temperature:   req.Payload.Temperature,
		MaxTokens:     int(req.Payload.MaxTokens),
		TopP:          req.Payload.TopP,
		StopSequences: req.Payload.Stop,
	}

	// Execute completion
	response, fromCache, err := c.usecase.Complete(ctx, req.Payload.ModelId, providerReq)
	if err != nil {
		return &aiproxy.CompleteResponse{
			Result: &aiproxy.Result{
				Code:    aiproxy.ResultCode_INTERNAL,
				Message: err.Error(),
			},
		}, nil
	}

	// Convert provider response to proto response
	return &aiproxy.CompleteResponse{
		Result: &aiproxy.Result{
			Code:    aiproxy.ResultCode_SUCCESS,
			Message: "Success",
		},
		Completion: &aiproxy.CompletionResponse{
			Id:         "", // TODO: generate ID
			ModelId:    req.Payload.ModelId,
			Text:       response.Content,
			TokensUsed: int64(response.TokensUsed),
			LatencyMs:  0, // TODO: track latency
			FromCache:  fromCache,
			Provider:   "", // TODO: track provider
		},
	}, nil
}

// StreamComplete handles streaming completion requests (not implemented yet)
func (c *ProxyController) StreamComplete(req *aiproxy.CompleteRequest, stream aiproxy.AIProxyService_StreamCompleteServer) error {
	// TODO: Implement streaming
	return nil
}

// HealthCheck handles health check requests
func (c *ProxyController) HealthCheck(ctx context.Context, req *aiproxy.HealthCheckRequest) (*aiproxy.HealthCheckResponse, error) {
	healthy, err := c.usecase.HealthCheck(ctx)
	if err != nil {
		return &aiproxy.HealthCheckResponse{
			Result: &aiproxy.Result{
				Code:    aiproxy.ResultCode_INTERNAL,
				Message: err.Error(),
			},
		}, nil
	}

	status := "unhealthy"
	if healthy {
		status = "healthy"
	}

	return &aiproxy.HealthCheckResponse{
		Result: &aiproxy.Result{
			Code:    aiproxy.ResultCode_SUCCESS,
			Message: "Success",
		},
		Status: status,
	}, nil
}

// GetProviderStatus handles provider status requests
func (c *ProxyController) GetProviderStatus(ctx context.Context, req *aiproxy.GetProviderStatusRequest) (*aiproxy.GetProviderStatusResponse, error) {
	// TODO: Implement provider status tracking
	return &aiproxy.GetProviderStatusResponse{
		Result: &aiproxy.Result{
			Code:    aiproxy.ResultCode_SUCCESS,
			Message: "Success",
		},
		Providers: []*aiproxy.ProviderHealth{},
	}, nil
}
