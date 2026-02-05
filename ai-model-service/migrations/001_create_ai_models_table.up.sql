-- Create AI models table
CREATE TABLE IF NOT EXISTS ai_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    provider VARCHAR(100) NOT NULL,
    model_id VARCHAR(255) NOT NULL,
    base_url TEXT NOT NULL,
    vault_path VARCHAR(255) NOT NULL,
    config JSONB DEFAULT '{}',
    quota_daily BIGINT DEFAULT 100000,
    quota_monthly BIGINT DEFAULT 3000000,
    cost_per_1k_tokens DECIMAL(10,4),
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_models_name ON ai_models(name);
CREATE INDEX IF NOT EXISTS idx_models_provider ON ai_models(provider);
CREATE INDEX IF NOT EXISTS idx_models_status ON ai_models(status);

-- Add comment
COMMENT ON TABLE ai_models IS 'AI model configurations with Vault integration';
