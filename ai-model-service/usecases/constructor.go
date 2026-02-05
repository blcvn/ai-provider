package usecases

import "github.com/blcvn/backend/services/ai-model-service/helper"

// NewModelUsecase creates a new model usecase
// NewModelUsecase creates a new model usecase
func NewModelUsecase(
	repository iModelRepository,
	vaultClient iVaultClient,
	crypto helper.CryptoHelpers,
) *modelUsecase {
	return &modelUsecase{
		repository:  repository,
		vaultClient: vaultClient,
		crypto:      crypto,
	}
}
