package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/blcvn/backend/services/ai-model-service/common/constants"
	"github.com/blcvn/backend/services/ai-model-service/common/errors"
	"github.com/blcvn/backend/services/ai-model-service/dto"
	"github.com/blcvn/backend/services/ai-model-service/entities"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type modelRepository struct {
	db *gorm.DB
}

// NewModelRepository creates a new model repository
func NewModelRepository(db *gorm.DB) *modelRepository {
	return &modelRepository{db: db}
}

// CreateModel creates a new AI model
func (r *modelRepository) CreateModel(ctx context.Context, payload *entities.CreateModelPayload, encryptedAPIKey string) (*entities.AIModel, errors.BaseError) {
	// Check if model with same name exists
	var existing dto.AIModel
	if err := r.db.Where("name = ?", payload.Name).First(&existing).Error; err == nil {
		return nil, errors.Conflict(constants.ErrModelAlreadyExists)
	}

	// Marshal config to JSON
	configJSON, _ := json.Marshal(payload.Config)

	// Create DTO
	dtoModel := &dto.AIModel{
		ID:              uuid.New(),
		Name:            payload.Name,
		Provider:        payload.Provider,
		ModelID:         payload.ModelID,
		BaseURL:         payload.BaseURL,
		EncryptedAPIKey: encryptedAPIKey,
		Config:          string(configJSON),
		QuotaDaily:      payload.QuotaDaily,
		QuotaMonthly:    payload.QuotaMonthly,
		CostPer1kTokens: payload.CostPer1kTokens,
		Status:          string(entities.ModelStatusActive),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Save to database
	if err := r.db.WithContext(ctx).Create(dtoModel).Error; err != nil {
		return nil, errors.Internal(fmt.Errorf(constants.ErrFailedToCreateModel, err))
	}

	// Convert to entity
	return r.dtoToEntity(dtoModel)
}

// GetModel retrieves a model by ID
func (r *modelRepository) GetModel(ctx context.Context, id string) (*entities.AIModel, errors.BaseError) {
	modelUUID, err := uuid.Parse(id)
	if err != nil {
		// If not a UUID, try to find by name
		return r.GetModelByName(ctx, id)
	}

	var dtoModel dto.AIModel
	if err := r.db.WithContext(ctx).Where("id = ?", modelUUID).First(&dtoModel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound(constants.ErrModelNotFound)
		}
		return nil, errors.Internal(err)
	}

	return r.dtoToEntity(&dtoModel)
}

// GetModelByName retrieves a model by name
func (r *modelRepository) GetModelByName(ctx context.Context, name string) (*entities.AIModel, errors.BaseError) {
	var dtoModel dto.AIModel
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&dtoModel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound(constants.ErrModelNotFound)
		}
		return nil, errors.Internal(err)
	}

	return r.dtoToEntity(&dtoModel)
}

// ListModels lists models with filtering
func (r *modelRepository) ListModels(ctx context.Context, filter *entities.ModelFilter) ([]*entities.AIModel, int64, errors.BaseError) {
	query := r.db.WithContext(ctx).Model(&dto.AIModel{})

	// Apply filters
	if filter.Provider != "" {
		query = query.Where("provider = ?", filter.Provider)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", string(filter.Status))
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Internal(err)
	}

	// Apply pagination
	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(int(offset)).Limit(int(filter.PageSize))
	}

	// Fetch models
	var dtoModels []dto.AIModel
	if err := query.Order("created_at DESC").Find(&dtoModels).Error; err != nil {
		return nil, 0, errors.Internal(err)
	}

	// Convert to entities
	models := make([]*entities.AIModel, 0, len(dtoModels))
	for i := range dtoModels {
		model, err := r.dtoToEntity(&dtoModels[i])
		if err != nil {
			continue
		}
		models = append(models, model)
	}

	return models, total, nil
}

// UpdateModel updates an existing model
func (r *modelRepository) UpdateModel(ctx context.Context, payload *entities.UpdateModelPayload) (*entities.AIModel, errors.BaseError) {
	modelUUID, err := uuid.Parse(payload.ID)
	if err != nil {
		return nil, errors.BadRequest(constants.ErrInvalidModelID)
	}

	// Check if model exists
	var dtoModel dto.AIModel
	if err := r.db.WithContext(ctx).Where("id = ?", modelUUID).First(&dtoModel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound(constants.ErrModelNotFound)
		}
		return nil, errors.Internal(err)
	}

	// Prepare updates
	updates := make(map[string]interface{})
	if payload.Name != "" {
		updates["name"] = payload.Name
	}
	if payload.Config != nil {
		configJSON, _ := json.Marshal(payload.Config)
		updates["config"] = string(configJSON)
	}
	if payload.QuotaDaily > 0 {
		updates["quota_daily"] = payload.QuotaDaily
	}
	if payload.QuotaMonthly > 0 {
		updates["quota_monthly"] = payload.QuotaMonthly
	}
	if payload.Status != "" {
		updates["status"] = string(payload.Status)
	}
	updates["updated_at"] = time.Now()

	// Update
	if err := r.db.WithContext(ctx).Model(&dtoModel).Updates(updates).Error; err != nil {
		return nil, errors.Internal(fmt.Errorf(constants.ErrFailedToUpdateModel, err))
	}

	// Fetch updated model
	return r.GetModel(ctx, payload.ID)
}

// DeleteModel deletes a model
func (r *modelRepository) DeleteModel(ctx context.Context, id string) errors.BaseError {
	modelUUID, err := uuid.Parse(id)
	if err != nil {
		return errors.BadRequest(constants.ErrInvalidModelID)
	}

	result := r.db.WithContext(ctx).Delete(&dto.AIModel{}, "id = ?", modelUUID)
	if result.Error != nil {
		return errors.Internal(fmt.Errorf(constants.ErrFailedToDeleteModel, result.Error))
	}
	if result.RowsAffected == 0 {
		return errors.NotFound(constants.ErrModelNotFound)
	}

	return nil
}

// LogUsage logs AI usage
func (r *modelRepository) LogUsage(ctx context.Context, payload *entities.LogUsagePayload) errors.BaseError {
	modelUUID, err := uuid.Parse(payload.ModelID)
	if err != nil {
		return errors.BadRequest(constants.ErrInvalidModelID)
	}

	var userUUID, sessionUUID *uuid.UUID
	if payload.UserID != "" {
		parsed, _ := uuid.Parse(payload.UserID)
		userUUID = &parsed
	}
	if payload.SessionID != "" {
		parsed, _ := uuid.Parse(payload.SessionID)
		sessionUUID = &parsed
	}

	// Get model to calculate cost
	var model dto.AIModel
	if err := r.db.WithContext(ctx).Where("id = ?", modelUUID).First(&model).Error; err != nil {
		return errors.NotFound(constants.ErrModelNotFound)
	}

	// Calculate cost
	cost := model.CostPer1kTokens.Mul(decimal.NewFromInt(payload.TokensUsed)).Div(decimal.NewFromInt(1000))

	// Create usage log
	usageLog := &dto.UsageLog{
		ID:           uuid.New(),
		ModelID:      modelUUID,
		UserID:       userUUID,
		SessionID:    sessionUUID,
		PromptHash:   payload.PromptHash,
		TokensUsed:   payload.TokensUsed,
		Cost:         cost,
		LatencyMs:    payload.LatencyMs,
		Status:       string(payload.Status),
		ErrorMessage: payload.ErrorMessage,
		CreatedAt:    time.Now(),
	}

	if err := r.db.WithContext(ctx).Create(usageLog).Error; err != nil {
		return errors.Internal(fmt.Errorf(constants.ErrFailedToLogUsage, err))
	}

	return nil
}

// GetDailyUsage gets total tokens used today for a model
func (r *modelRepository) GetDailyUsage(ctx context.Context, modelID string, date time.Time) (int64, errors.BaseError) {
	modelUUID, err := uuid.Parse(modelID)
	if err != nil {
		return 0, errors.BadRequest(constants.ErrInvalidModelID)
	}

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var total int64
	if err := r.db.WithContext(ctx).Model(&dto.UsageLog{}).
		Where("model_id = ? AND created_at >= ? AND created_at < ? AND status = ?",
			modelUUID, startOfDay, endOfDay, string(entities.UsageStatusSuccess)).
		Select("COALESCE(SUM(tokens_used), 0)").
		Scan(&total).Error; err != nil {
		return 0, errors.Internal(fmt.Errorf(constants.ErrFailedToGetUsage, err))
	}

	return total, nil
}

// GetMonthlyUsage gets total tokens used this month for a model
func (r *modelRepository) GetMonthlyUsage(ctx context.Context, modelID string, year int, month int) (int64, errors.BaseError) {
	modelUUID, err := uuid.Parse(modelID)
	if err != nil {
		return 0, errors.BadRequest(constants.ErrInvalidModelID)
	}

	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	var total int64
	if err := r.db.WithContext(ctx).Model(&dto.UsageLog{}).
		Where("model_id = ? AND created_at >= ? AND created_at < ? AND status = ?",
			modelUUID, startOfMonth, endOfMonth, string(entities.UsageStatusSuccess)).
		Select("COALESCE(SUM(tokens_used), 0)").
		Scan(&total).Error; err != nil {
		return 0, errors.Internal(fmt.Errorf(constants.ErrFailedToGetUsage, err))
	}

	return total, nil
}

// Helper: Convert DTO to Entity
func (r *modelRepository) dtoToEntity(dtoModel *dto.AIModel) (*entities.AIModel, errors.BaseError) {
	var config map[string]string
	if err := json.Unmarshal([]byte(dtoModel.Config), &config); err != nil {
		config = make(map[string]string)
	}

	return &entities.AIModel{
		ID:              dtoModel.ID.String(),
		Name:            dtoModel.Name,
		Provider:        dtoModel.Provider,
		ModelID:         dtoModel.ModelID,
		BaseURL:         dtoModel.BaseURL,
		EncryptedAPIKey: dtoModel.EncryptedAPIKey,
		Config:          config,
		QuotaDaily:      dtoModel.QuotaDaily,
		QuotaMonthly:    dtoModel.QuotaMonthly,
		CostPer1kTokens: dtoModel.CostPer1kTokens,
		Status:          entities.ModelStatus(dtoModel.Status),
		CreatedAt:       dtoModel.CreatedAt,
		UpdatedAt:       dtoModel.UpdatedAt,
	}, nil
}
