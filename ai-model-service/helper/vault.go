package helper

import (
	"context"
	"fmt"
	"time"

	"github.com/blcvn/backend/services/ai-model-service/common/constants"
	"github.com/blcvn/backend/services/ai-model-service/common/errors"
	"github.com/blcvn/backend/services/ai-model-service/entities"
	vault "github.com/hashicorp/vault/api"
	"github.com/jellydator/ttlcache/v3"
)

// VaultClient wraps Vault API client with caching
type VaultClient struct {
	client *vault.Client
	cache  *ttlcache.Cache[string, *entities.Credentials]
}

// NewVaultClient creates a new Vault client with caching
func NewVaultClient(addr, token string) (*VaultClient, errors.BaseError) {
	config := vault.DefaultConfig()
	config.Address = addr

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, errors.Internal(fmt.Errorf(constants.ErrVaultConnectionFailed, err))
	}

	client.SetToken(token)

	// Create cache with 5-minute TTL
	cache := ttlcache.New[string, *entities.Credentials](
		ttlcache.WithTTL[string, *entities.Credentials](5 * time.Minute),
	)
	go cache.Start()

	return &VaultClient{
		client: client,
		cache:  cache,
	}, nil
}

// GetCredentials retrieves credentials from Vault with caching
func (v *VaultClient) GetCredentials(ctx context.Context, vaultPath string) (*entities.Credentials, errors.BaseError) {
	// Check cache first
	if item := v.cache.Get(vaultPath); item != nil {
		return item.Value(), nil
	}

	// Read from Vault
	secret, err := v.client.Logical().ReadWithContext(ctx, vaultPath)
	if err != nil {
		return nil, errors.Internal(fmt.Errorf(constants.ErrVaultReadFailed, err))
	}

	if secret == nil || secret.Data == nil {
		return nil, errors.NotFound(fmt.Sprintf(constants.ErrVaultPathNotFound, vaultPath))
	}

	// Extract credentials
	apiKey, ok := secret.Data["api_key"].(string)
	if !ok {
		return nil, errors.Internal(fmt.Errorf(constants.ErrInvalidCredentials))
	}

	baseURL, _ := secret.Data["base_url"].(string)

	// Parse headers if present
	headers := make(map[string]string)
	if headersData, ok := secret.Data["headers"].(map[string]interface{}); ok {
		for k, v := range headersData {
			if strVal, ok := v.(string); ok {
				headers[k] = strVal
			}
		}
	}

	creds := &entities.Credentials{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Headers: headers,
	}

	// Cache the credentials
	v.cache.Set(vaultPath, creds, ttlcache.DefaultTTL)

	return creds, nil
}

// Close stops the cache cleanup goroutine
func (v *VaultClient) Close() {
	v.cache.Stop()
}
