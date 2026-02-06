package helper

import (
	"context"
	"fmt"

	model_pb "github.com/blcvn/kratos-proto/go/ai-model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AIModelClient struct {
	client model_pb.AIModelServiceClient
}

func NewAIModelClient(addr string) (*AIModelClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &AIModelClient{
		client: model_pb.NewAIModelServiceClient(conn),
	}, nil
}

func (c *AIModelClient) GetCredentials(ctx context.Context, modelID string) (*model_pb.Credentials, error) {
	resp, err := c.client.GetCredentials(ctx, &model_pb.GetCredentialsRequest{ModelId: modelID})
	if err != nil {
		return nil, err
	}
	if resp.Result.Code != model_pb.ResultCode_SUCCESS {
		return nil, fmt.Errorf("failed to get credentials: %s", resp.Result.Message)
	}
	return resp.Credentials, nil
}

func (c *AIModelClient) GetModel(ctx context.Context, modelID string) (*model_pb.AIModel, error) {
	resp, err := c.client.GetModel(ctx, &model_pb.GetModelRequest{Id: modelID})
	if err != nil {
		return nil, err
	}
	if resp.Result.Code != model_pb.ResultCode_SUCCESS {
		return nil, fmt.Errorf("failed to get model: %s", resp.Result.Message)
	}
	return resp.Model, nil
}

func (c *AIModelClient) CheckQuota(ctx context.Context, modelID string, tokens int32) (bool, error) {
	// Note: CheckQuotaRequest in proto currently only supports model_id.
	// Tokens check might need proto update or logic change.
	// For now, we only check if model has quota status available.
	resp, err := c.client.CheckQuota(ctx, &model_pb.CheckQuotaRequest{ModelId: modelID})
	if err != nil {
		return false, err
	}
	// CheckQuotaResponse has QuotaStatus which has boolean Exceeded? No, it has QuotaStatus.
	// But resp.Result can be checked too.
	// Helper should probably return QuotaStatus or bool based on logic.
	// Assuming QuotaStatus has Exceeded field as per base.proto
	if resp.Quota != nil {
		return !resp.Quota.Exceeded, nil
	}
	return true, nil
}

func (c *AIModelClient) LogUsage(ctx context.Context, modelID string, promptTokens, completionTokens int32) error {
	_, err := c.client.LogUsage(ctx, &model_pb.LogUsageRequest{
		Payload: &model_pb.LogUsagePayload{
			ModelId:    modelID,
			TokensUsed: int64(promptTokens + completionTokens),
		},
	})
	return err
}
