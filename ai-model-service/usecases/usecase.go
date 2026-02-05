package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/blcvn/backend/services/ai-model-service/common/constants"
	"github.com/blcvn/backend/services/ai-model-service/common/errors"
	"github.com/blcvn/backend/services/ai-model-service/entities"
	"github.com/blcvn/backend/services/ai-model-service/helper"
)

type modelUsecase struct {
	repository  iModelRepository
	vaultClient iVaultClient
	crypto      helper.CryptoHelpers
}

// CreateModel creates a new AI model
func (u *modelUsecase) CreateModel(ctx context.Context, payload *entities.CreateModelPayload) (*entities.AIModel, errors.BaseError) {
	// Validate payload
	if payload.Name == "" {
		return nil, errors.BadRequest(constants.ErrInvalidModelName)
	}
	if payload.Provider == "" {
		return nil, errors.BadRequest(constants.ErrInvalidProvider)
	}
	if payload.APIKey == "" {
		return nil, errors.BadRequest("api key is required")
	}

	// Encrypt API Key
	encryptedKey, cryptoErr := u.crypto.Encrypt(payload.APIKey)
	if cryptoErr != nil {
		return nil, errors.Internal(fmt.Errorf("failed to encrypt api key: %v", cryptoErr))
	}

	// Create model via repository
	model, repoErr := u.repository.CreateModel(ctx, payload, encryptedKey)
	if repoErr != nil {
		return nil, repoErr
	}

	return model, nil
}

// GetModel retrieves a model by ID
func (u *modelUsecase) GetModel(ctx context.Context, id string) (*entities.AIModel, errors.BaseError) {
	return u.repository.GetModel(ctx, id)
}

// ListModels lists models with filtering
func (u *modelUsecase) ListModels(ctx context.Context, filter *entities.ModelFilter) ([]*entities.AIModel, int64, errors.BaseError) {
	return u.repository.ListModels(ctx, filter)
}

// UpdateModel updates an existing model
func (u *modelUsecase) UpdateModel(ctx context.Context, payload *entities.UpdateModelPayload) (*entities.AIModel, errors.BaseError) {
	// Validate payload
	if payload.ID == "" {
		return nil, errors.BadRequest(constants.ErrInvalidModelID)
	}

	return u.repository.UpdateModel(ctx, payload)
}

// DeleteModel deletes a model
func (u *modelUsecase) DeleteModel(ctx context.Context, id string) errors.BaseError {
	return u.repository.DeleteModel(ctx, id)
}

// GetCredentials retrieves API credentials from Vault
func (u *modelUsecase) GetCredentials(ctx context.Context, modelID string) (*entities.Credentials, errors.BaseError) {
	// Get model to retrieve vault path
	model, err := u.repository.GetModel(ctx, modelID)
	if err != nil {
		return nil, err
	}

	// Check if model is active
	if model.Status != entities.ModelStatusActive {
		return nil, errors.BadRequest(fmt.Sprintf("model is not active: %s", model.Status))
	}

	// Decrypt API Key
	apiKey, errStr := u.crypto.Decrypt(model.EncryptedAPIKey)
	if errStr != nil {
		return nil, errors.Internal(fmt.Errorf("failed to decrypt api key: %v", errStr))
	}

	// Construct credentials
	creds := &entities.Credentials{
		APIKey:  apiKey,
		BaseURL: model.BaseURL,
		Headers: make(map[string]string), // Initialize empty map or load from config if needed
	}

	return creds, nil
}

// LogUsage logs AI usage
func (u *modelUsecase) LogUsage(ctx context.Context, payload *entities.LogUsagePayload) errors.BaseError {
	// Validate model exists
	_, err := u.repository.GetModel(ctx, payload.ModelID)
	if err != nil {
		return err
	}

	return u.repository.LogUsage(ctx, payload)
}

// CheckQuota checks if quota is exceeded
func (u *modelUsecase) CheckQuota(ctx context.Context, modelID string) (*entities.QuotaStatus, errors.BaseError) {
	// Get model
	model, err := u.repository.GetModel(ctx, modelID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// Get daily usage
	dailyUsed, err := u.repository.GetDailyUsage(ctx, modelID, now)
	if err != nil {
		return nil, err
	}

	// Get monthly usage
	monthlyUsed, err := u.repository.GetMonthlyUsage(ctx, modelID, now.Year(), int(now.Month()))
	if err != nil {
		return nil, err
	}

	// Check if exceeded
	exceeded := false
	if model.QuotaDaily > 0 && dailyUsed >= model.QuotaDaily {
		exceeded = true
	}
	if model.QuotaMonthly > 0 && monthlyUsed >= model.QuotaMonthly {
		exceeded = true
	}

	// Calculate reset time (next day at midnight)
	tomorrow := now.AddDate(0, 0, 1)
	resetTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())

	return &entities.QuotaStatus{
		Exceeded:     exceeded,
		DailyUsed:    dailyUsed,
		DailyLimit:   model.QuotaDaily,
		MonthlyUsed:  monthlyUsed,
		MonthlyLimit: model.QuotaMonthly,
		ResetTime:    resetTime.Format(time.RFC3339),
	}, nil
}
