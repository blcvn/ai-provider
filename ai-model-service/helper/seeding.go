package helper

import (
	"log"
	"time"

	"github.com/blcvn/backend/services/ai-model-service/dto"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ModelSeed struct {
	Name            string
	Provider        string
	ModelID         string
	BaseURL         string
	CostPer1kTokens decimal.Decimal
}

func SeedModels(db *gorm.DB, crypto CryptoHelpers) {
	apiKey := "DUMMY_KEY"
	encryptedKey, err := crypto.Encrypt(apiKey)
	if err != nil {
		log.Printf("ERROR: Failed to encrypt API key for seeding: %v", err)
		return
	}

	models := []ModelSeed{
		{
			Name:            "claude-haiku-4-5-20251001",
			Provider:        "anthropic",
			ModelID:         "claude-haiku-4-5-20251001",
			BaseURL:         "https://api.anthropic.com/v1",
			CostPer1kTokens: decimal.NewFromFloat(0.00025),
		},
		{
			Name:            "claude-sonnet-4-5-20250929",
			Provider:        "anthropic",
			ModelID:         "claude-sonnet-4-5-20250929",
			BaseURL:         "https://api.anthropic.com/v1",
			CostPer1kTokens: decimal.NewFromFloat(0.003),
		},
		{
			Name:            "claude-opus-4-5-20251101",
			Provider:        "anthropic",
			ModelID:         "claude-opus-4-5-20251101",
			BaseURL:         "https://api.anthropic.com/v1",
			CostPer1kTokens: decimal.NewFromFloat(0.015),
		},
		{
			Name:            "claude-3-opus-20240229",
			Provider:        "anthropic",
			ModelID:         "claude-3-opus-20240229",
			BaseURL:         "https://api.anthropic.com/v1",
			CostPer1kTokens: decimal.NewFromFloat(0.015),
		},
	}

	for _, m := range models {
		var existing dto.AIModel
		err := db.Where("name = ?", m.Name).First(&existing).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create new
				newModel := dto.AIModel{
					ID:              uuid.New(),
					Name:            m.Name,
					Provider:        m.Provider,
					ModelID:         m.ModelID,
					BaseURL:         m.BaseURL,
					EncryptedAPIKey: encryptedKey,
					CostPer1kTokens: m.CostPer1kTokens,
					Status:          "active",
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
				}
				if err := db.Create(&newModel).Error; err != nil {
					log.Printf("ERROR: Failed to seed model %s: %v", m.Name, err)
				} else {
					log.Printf("INFO: Seeded new model: %s", m.Name)
				}
			} else {
				log.Printf("ERROR: Failed to check model %s: %v", m.Name, err)
			}
			continue
		}

		// Update existing
		updates := map[string]interface{}{
			"encrypted_api_key": encryptedKey,
			"model_id":          m.ModelID,
			"provider":          m.Provider,
			"base_url":          m.BaseURL,
			"cost_per1k_tokens": m.CostPer1kTokens,
			"status":            "active",
			"updated_at":        time.Now(),
		}
		if err := db.Model(&existing).Updates(updates).Error; err != nil {
			log.Printf("ERROR: Failed to update model %s: %v", m.Name, err)
		} else {
			log.Printf("INFO: Updated seeded model: %s", m.Name)
		}
	}
}
