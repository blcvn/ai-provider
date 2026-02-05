package helper

import (
	"fmt"

	"github.com/blcvn/backend/services/ai-model-service/entities"
	pb "github.com/blcvn/kratos-proto/go/ai-model"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Transform struct{}

// NewTransform creates a new transform helper
func NewTransform() *Transform {
	return &Transform{}
}

// Model2Pb converts entity to proto
func (t *Transform) Model2Pb(model *entities.AIModel) (*pb.AIModel, error) {
	if model == nil {
		return nil, fmt.Errorf("model is nil")
	}

	costFloat, _ := model.CostPer1kTokens.Float64()

	return &pb.AIModel{
		Id:       model.ID,
		Name:     model.Name,
		Provider: model.Provider,
		ModelId:  model.ModelID,
		BaseUrl:  model.BaseURL,
		// VaultPath removed
		Config:           model.Config,
		QuotaDaily:       model.QuotaDaily,
		QuotaMonthly:     model.QuotaMonthly,
		CostPer_1KTokens: costFloat,
		Status:           pb.ModelStatus(pb.ModelStatus_value[string(model.Status)]),
		CreatedAt:        timestamppb.New(model.CreatedAt),
		UpdatedAt:        timestamppb.New(model.UpdatedAt),
	}, nil
}

// Credentials2Pb converts entity to proto
func (t *Transform) Credentials2Pb(creds *entities.Credentials) (*pb.Credentials, error) {
	if creds == nil {
		return nil, fmt.Errorf("credentials is nil")
	}

	return &pb.Credentials{
		ApiKey:  creds.APIKey,
		BaseUrl: creds.BaseURL,
		Headers: creds.Headers,
	}, nil
}

// QuotaStatus2Pb converts entity to proto
func (t *Transform) QuotaStatus2Pb(quota *entities.QuotaStatus) (*pb.QuotaStatus, error) {
	if quota == nil {
		return nil, fmt.Errorf("quota is nil")
	}

	return &pb.QuotaStatus{
		Exceeded:     quota.Exceeded,
		DailyUsed:    quota.DailyUsed,
		DailyLimit:   quota.DailyLimit,
		MonthlyUsed:  quota.MonthlyUsed,
		MonthlyLimit: quota.MonthlyLimit,
		ResetTime:    quota.ResetTime,
	}, nil
}

// Pb2CreateModelPayload converts proto to entity
func (t *Transform) Pb2CreateModelPayload(pb *pb.CreateModelPayload) (*entities.CreateModelPayload, error) {
	if pb == nil {
		return nil, fmt.Errorf("payload is nil")
	}

	return &entities.CreateModelPayload{
		Name:            pb.Name,
		Provider:        pb.Provider,
		ModelID:         pb.ModelId,
		BaseURL:         pb.BaseUrl,
		APIKey:          pb.ApiKey,
		Config:          pb.Config,
		QuotaDaily:      pb.QuotaDaily,
		QuotaMonthly:    pb.QuotaMonthly,
		CostPer1kTokens: decimal.NewFromFloat(pb.CostPer_1KTokens),
	}, nil
}

// Pb2UpdateModelPayload converts proto to entity
func (t *Transform) Pb2UpdateModelPayload(pb *pb.UpdateModelPayload) (*entities.UpdateModelPayload, error) {
	if pb == nil {
		return nil, fmt.Errorf("payload is nil")
	}

	return &entities.UpdateModelPayload{
		ID:           pb.Id,
		Name:         pb.Name,
		Config:       pb.Config,
		QuotaDaily:   pb.QuotaDaily,
		QuotaMonthly: pb.QuotaMonthly,
		Status:       entities.ModelStatus(pb.Status.String()),
	}, nil
}

// Pb2LogUsagePayload converts proto to entity
func (t *Transform) Pb2LogUsagePayload(pb *pb.LogUsagePayload) (*entities.LogUsagePayload, error) {
	if pb == nil {
		return nil, fmt.Errorf("payload is nil")
	}

	return &entities.LogUsagePayload{
		ModelID:      pb.ModelId,
		UserID:       pb.UserId,
		SessionID:    pb.SessionId,
		PromptHash:   pb.PromptHash,
		TokensUsed:   pb.TokensUsed,
		LatencyMs:    pb.LatencyMs,
		Status:       entities.UsageStatus(pb.Status.String()),
		ErrorMessage: pb.ErrorMessage,
	}, nil
}

// Pb2ModelFilter converts proto to entity
func (t *Transform) Pb2ModelFilter(provider string, status pb.ModelStatus, page, pageSize int32) (*entities.ModelFilter, error) {
	return &entities.ModelFilter{
		Provider: provider,
		Status:   entities.ModelStatus(status.String()),
		Page:     page,
		PageSize: pageSize,
	}, nil
}
