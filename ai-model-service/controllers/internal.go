package controllers

import (
	"context"

	"github.com/blcvn/backend/services/ai-model-service/common/errors"
	"github.com/blcvn/backend/services/ai-model-service/entities"
	pb "github.com/blcvn/kratos-proto/go/ai-model"
)

// iModelUsecase defines usecase interface
type iModelUsecase interface {
	CreateModel(ctx context.Context, payload *entities.CreateModelPayload) (*entities.AIModel, errors.BaseError)
	GetModel(ctx context.Context, id string) (*entities.AIModel, errors.BaseError)
	ListModels(ctx context.Context, filter *entities.ModelFilter) ([]*entities.AIModel, int64, errors.BaseError)
	UpdateModel(ctx context.Context, payload *entities.UpdateModelPayload) (*entities.AIModel, errors.BaseError)
	DeleteModel(ctx context.Context, id string) errors.BaseError
	GetCredentials(ctx context.Context, modelID string) (*entities.Credentials, errors.BaseError)
	LogUsage(ctx context.Context, payload *entities.LogUsagePayload) errors.BaseError
	CheckQuota(ctx context.Context, modelID string) (*entities.QuotaStatus, errors.BaseError)
}

// iTransform defines transformation interface
type iTransform interface {
	// Entity to Proto
	Model2Pb(model *entities.AIModel) (*pb.AIModel, error)
	Credentials2Pb(creds *entities.Credentials) (*pb.Credentials, error)
	QuotaStatus2Pb(quota *entities.QuotaStatus) (*pb.QuotaStatus, error)

	// Proto to Entity
	Pb2CreateModelPayload(pb *pb.CreateModelPayload) (*entities.CreateModelPayload, error)
	Pb2UpdateModelPayload(pb *pb.UpdateModelPayload) (*entities.UpdateModelPayload, error)
	Pb2LogUsagePayload(pb *pb.LogUsagePayload) (*entities.LogUsagePayload, error)
	Pb2ModelFilter(provider string, status pb.ModelStatus, page, pageSize int32) (*entities.ModelFilter, error)
}
