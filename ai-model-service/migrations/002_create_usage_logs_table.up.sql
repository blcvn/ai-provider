-- Create AI usage logs table
CREATE TABLE IF NOT EXISTS ai_usage_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES ai_models(id) ON DELETE CASCADE,
    user_id UUID,
    session_id UUID,
    prompt_hash VARCHAR(64),
    tokens_used BIGINT NOT NULL,
    cost DECIMAL(10,4),
    latency_ms INTEGER,
    status VARCHAR(50) NOT NULL,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_usage_model_created ON ai_usage_logs(model_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_usage_user_created ON ai_usage_logs(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_usage_prompt_hash ON ai_usage_logs(prompt_hash);
CREATE INDEX IF NOT EXISTS idx_usage_created_date ON ai_usage_logs(DATE(created_at));

-- Add comment
COMMENT ON TABLE ai_usage_logs IS 'AI usage tracking for quota management and analytics';
