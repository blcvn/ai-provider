-- Migration to replace vault_path with encrypted_api_key
ALTER TABLE ai_models DROP COLUMN IF EXISTS vault_path;
ALTER TABLE ai_models ADD COLUMN encrypted_api_key TEXT;
