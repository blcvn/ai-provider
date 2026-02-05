package usecases

import (
	"context"
	"time"

	"github.com/blcvn/backend/services/ai-model-service/common/errors"
	"github.com/blcvn/backend/services/ai-model-service/entities"
)

// iModelRepository defines repository interface
type iModelRepository interface {
	CreateModel(ctx context.Context, payload *entities.CreateModelPayload, encryptedAPIKey string) (*entities.AIModel, errors.BaseError)
	GetModel(ctx context.Context, id string) (*entities.AIModel, errors.BaseError)
	GetModelByName(ctx context.Context, name string) (*entities.AIModel, errors.BaseError)
	ListModels(ctx context.Context, filter *entities.ModelFilter) ([]*entities.AIModel, int64, errors.BaseError)
	UpdateModel(ctx context.Context, payload *entities.UpdateModelPayload) (*entities.AIModel, errors.BaseError)
	DeleteModel(ctx context.Context, id string) errors.BaseError
	LogUsage(ctx context.Context, payload *entities.LogUsagePayload) errors.BaseError
	GetDailyUsage(ctx context.Context, modelID string, date time.Time) (int64, errors.BaseError)
	GetMonthlyUsage(ctx context.Context, modelID string, year int, month int) (int64, errors.BaseError)
}

// iVaultClient defines Vault client interface
type iVaultClient interface {
	GetCredentials(ctx context.Context, vaultPath string) (*entities.Credentials, errors.BaseError)
}
