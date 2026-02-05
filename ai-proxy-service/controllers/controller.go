package controllers

import (
	"context"
	"strings"

	"github.com/blcvn/backend/services/ai-proxy-service/common/errors"
	"github.com/blcvn/backend/services/ai-proxy-service/entities"
	pb "github.com/blcvn/kratos-proto/go/ai-proxy"
)

type iAIProxyUsecase interface {
	Complete(ctx context.Context, req *entities.CompletionRequest) (*entities.CompletionResponse, errors.BaseError)
	StreamComplete(ctx context.Context, req *entities.CompletionRequest, callback func(*entities.StreamResponse) error) errors.BaseError
}

type aiProxyController struct {
	pb.UnimplementedAIProxyServiceServer
	usecase iAIProxyUsecase
}

func NewAIProxyController(usecase iAIProxyUsecase) *aiProxyController {
	return &aiProxyController{usecase: usecase}
}

func (c *aiProxyController) Complete(ctx context.Context, req *pb.CompleteRequest) (*pb.CompleteResponse, error) {
	messages := make([]entities.Message, len(req.Payload.Messages))
	for i, m := range req.Payload.Messages {
		messages[i] = entities.Message{
			Role:    entities.MessageRole(strings.ToLower(m.Role.String())),
			Content: m.Content,
		}
	}

	entityReq := &entities.CompletionRequest{
		ModelID:     req.Payload.ModelId,
		Messages:    messages,
		Temperature: float32(req.Payload.Temperature),
		MaxTokens:   req.Payload.MaxTokens,
	}

	resp, err := c.usecase.Complete(ctx, entityReq)
	if err != nil {
		return &pb.CompleteResponse{
			Metadata: req.Metadata,
			Result:   &pb.Result{Code: pb.ResultCode(err.GetCode()), Message: err.Error()},
		}, nil
	}

	return &pb.CompleteResponse{
		Metadata: req.Metadata,
		Result:   &pb.Result{Code: pb.ResultCode_SUCCESS},
		Completion: &pb.CompletionResponse{
			Text:             resp.Content,
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

func (c *aiProxyController) StreamComplete(req *pb.CompleteRequest, stream pb.AIProxyService_StreamCompleteServer) error {
	messages := make([]entities.Message, len(req.Payload.Messages))
	for i, m := range req.Payload.Messages {
		messages[i] = entities.Message{
			Role:    entities.MessageRole(strings.ToLower(m.Role.String())),
			Content: m.Content,
		}
	}

	entityReq := &entities.CompletionRequest{
		ModelID:     req.Payload.ModelId,
		Messages:    messages,
		Temperature: float32(req.Payload.Temperature),
		MaxTokens:   req.Payload.MaxTokens,
	}

	err := c.usecase.StreamComplete(stream.Context(), entityReq, func(sr *entities.StreamResponse) error {
		pbResp := &pb.StreamCompleteResponse{
			Metadata: req.Metadata,
			Result:   &pb.Result{Code: pb.ResultCode_SUCCESS},
			Chunk: &pb.CompletionChunk{
				Text: sr.Content,
			},
		}
		// In streaming, we usually return usage only in the last chunk or via separate field
		// For now, mapping Content only as chunk text
		return stream.Send(pbResp)
	})

	return err
}
