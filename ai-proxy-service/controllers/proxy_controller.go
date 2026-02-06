package controllers

import (
	"context"

	"github.com/blcvn/backend/services/ai-proxy-service/entities"
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

	// Convert proto request to entity request
	entityReq := &entities.CompletionRequest{
		ModelID: req.Payload.ModelId,
		Messages: []entities.Message{
			{Role: entities.RoleUser, Content: req.Payload.Prompt}, // TODO: Support chat messages from proto if available or needed
		},
		Temperature:   float32(req.Payload.Temperature),
		MaxTokens:     int32(req.Payload.MaxTokens),
		StopSequences: req.Payload.Stop,
	}

	// Execute completion
	response, err := c.usecase.Complete(ctx, entityReq)
	fromCache := false // Caching temporarily removed in new usecase logic
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
			Id:               "", // TODO: generate ID
			ModelId:          req.Payload.ModelId,
			Text:             response.Content,
			TotalTokens:      int32(response.Usage.TotalTokens),
			PromptTokens:     int32(response.Usage.PromptTokens),
			CompletionTokens: int32(response.Usage.CompletionTokens),
			LatencyMs:        0, // TODO: track latency
			FromCache:        fromCache,
			Provider:         "", // TODO: track provider
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
