package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// AIModel represents the database model for AI models
type AIModel struct {
	ID              uuid.UUID       `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name            string          `gorm:"type:varchar(255);uniqueIndex;not null"`
	Provider        string          `gorm:"type:varchar(100);not null;index"`
	ModelID         string          `gorm:"type:varchar(255);not null"`
	BaseURL         string          `gorm:"type:text;not null"`
	EncryptedAPIKey string          `gorm:"type:text"`
	Config          string          `gorm:"type:jsonb;default:'{}'"`
	QuotaDaily      int64           `gorm:"default:100000"`
	QuotaMonthly    int64           `gorm:"default:3000000"`
	CostPer1kTokens decimal.Decimal `gorm:"type:decimal(10,4)"`
	Status          string          `gorm:"type:varchar(50);default:'active';index"`
	CreatedAt       time.Time       `gorm:"default:now()"`
	UpdatedAt       time.Time       `gorm:"default:now()"`
}

// TableName specifies the table name for AIModel
func (AIModel) TableName() string {
	return "ai_models"
}

// BeforeUpdate hook to update UpdatedAt
func (m *AIModel) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}

// UsageLog represents the database model for usage logs
type UsageLog struct {
	ID           uuid.UUID       `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ModelID      uuid.UUID       `gorm:"type:uuid;not null;index:idx_usage_model_created"`
	UserID       *uuid.UUID      `gorm:"type:uuid;index:idx_usage_user_created"`
	SessionID    *uuid.UUID      `gorm:"type:uuid"`
	PromptHash   string          `gorm:"type:varchar(64);index"`
	TokensUsed   int64           `gorm:"not null"`
	Cost         decimal.Decimal `gorm:"type:decimal(10,4)"`
	LatencyMs    int32           `gorm:"type:integer"`
	Status       string          `gorm:"type:varchar(50);not null"`
	ErrorMessage string          `gorm:"type:text"`
	CreatedAt    time.Time       `gorm:"default:now();index:idx_usage_model_created,idx_usage_user_created,idx_usage_created_date"`
}

// TableName specifies the table name for UsageLog
func (UsageLog) TableName() string {
	return "ai_usage_logs"
}
