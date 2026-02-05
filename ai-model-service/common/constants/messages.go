package constants

// Error messages
const (
	// Model errors
	ErrModelNotFound       = "model not found"
	ErrModelAlreadyExists  = "model with this name already exists"
	ErrInvalidModelID      = "invalid model ID format"
	ErrInvalidVaultPath    = "invalid vault path"
	ErrFailedToCreateModel = "failed to create model: %v"
	ErrFailedToUpdateModel = "failed to update model: %v"
	ErrFailedToDeleteModel = "failed to delete model: %v"

	// Vault errors
	ErrVaultConnectionFailed = "failed to connect to vault: %v"
	ErrVaultReadFailed       = "failed to read from vault: %v"
	ErrVaultPathNotFound     = "vault path not found: %s"
	ErrInvalidCredentials    = "invalid credentials from vault"

	// Quota errors
	ErrQuotaExceeded        = "quota exceeded for model: %s"
	ErrDailyQuotaExceeded   = "daily quota exceeded"
	ErrMonthlyQuotaExceeded = "monthly quota exceeded"
	ErrFailedToCheckQuota   = "failed to check quota: %v"

	// Usage errors
	ErrFailedToLogUsage = "failed to log usage: %v"
	ErrFailedToGetUsage = "failed to get usage: %v"

	// Validation errors
	ErrInvalidModelName = "model name is required"
	ErrInvalidProvider  = "provider is required"
	ErrInvalidBaseURL   = "base URL is required"
)

// Success messages
const (
	MsgModelCreated         = "model created successfully"
	MsgModelUpdated         = "model updated successfully"
	MsgModelDeleted         = "model deleted successfully"
	MsgModelRetrieved       = "model retrieved successfully"
	MsgModelsListed         = "models listed successfully"
	MsgCredentialsRetrieved = "credentials retrieved successfully"
	MsgUsageLogged          = "usage logged successfully"
	MsgQuotaChecked         = "quota checked successfully"
)
