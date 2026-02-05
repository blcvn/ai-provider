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

func (c *AIModelClient) CheckQuota(ctx context.Context, modelID string, tokens int32) (bool, error) {
	resp, err := c.client.CheckQuota(ctx, &model_pb.CheckQuotaRequest{ModelId: modelID, EstimatedTokens: tokens})
	if err != nil {
		return false, err
	}
	return resp.IsAllowed, nil
}

func (c *AIModelClient) LogUsage(ctx context.Context, modelID string, promptTokens, completionTokens int32) error {
	_, err := c.client.LogUsage(ctx, &model_pb.LogUsageRequest{
		ModelId:          modelID,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
	})
	return err
}
