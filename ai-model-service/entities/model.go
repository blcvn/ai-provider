package entities

import (
	"time"

	"github.com/shopspring/decimal"
)

// ModelStatus represents the status of an AI model
type ModelStatus string

const (
	ModelStatusActive     ModelStatus = "active"
	ModelStatusDisabled   ModelStatus = "disabled"
	ModelStatusDeprecated ModelStatus = "deprecated"
)

// AIModel represents an AI model configuration
type AIModel struct {
	ID              string
	Name            string
	Provider        string
	ModelID         string
	BaseURL         string
	EncryptedAPIKey string
	Config          map[string]string
	QuotaDaily      int64
	QuotaMonthly    int64
	CostPer1kTokens decimal.Decimal
	Status          ModelStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// UsageStatus represents the status of a usage log
type UsageStatus string

const (
	UsageStatusSuccess UsageStatus = "success"
	UsageStatusError   UsageStatus = "error"
	UsageStatusTimeout UsageStatus = "timeout"
)

// UsageLog represents an AI usage log entry
type UsageLog struct {
	ID           string
	ModelID      string
	UserID       string
	SessionID    string
	PromptHash   string
	TokensUsed   int64
	Cost         decimal.Decimal
	LatencyMs    int32
	Status       UsageStatus
	ErrorMessage string
	CreatedAt    time.Time
}

// Credentials represents API credentials from Vault
type Credentials struct {
	APIKey  string
	BaseURL string
	Headers map[string]string
}

// QuotaStatus represents quota usage information
type QuotaStatus struct {
	Exceeded     bool
	DailyUsed    int64
	DailyLimit   int64
	MonthlyUsed  int64
	MonthlyLimit int64
	ResetTime    string
}

// CreateModelPayload represents the payload for creating a model
type CreateModelPayload struct {
	Name            string
	Provider        string
	ModelID         string
	BaseURL         string
	APIKey          string
	Config          map[string]string
	QuotaDaily      int64
	QuotaMonthly    int64
	CostPer1kTokens decimal.Decimal
}

// UpdateModelPayload represents the payload for updating a model
type UpdateModelPayload struct {
	ID           string
	Name         string
	Config       map[string]string
	QuotaDaily   int64
	QuotaMonthly int64
	Status       ModelStatus
}

// LogUsagePayload represents the payload for logging usage
type LogUsagePayload struct {
	ModelID      string
	UserID       string
	SessionID    string
	PromptHash   string
	TokensUsed   int64
	LatencyMs    int32
	Status       UsageStatus
	ErrorMessage string
}

// ModelFilter represents filter criteria for listing models
type ModelFilter struct {
	Provider string
	Status   ModelStatus
	Page     int32
	PageSize int32
}
