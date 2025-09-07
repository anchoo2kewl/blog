CREATE TABLE api_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    token_hash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    CONSTRAINT api_tokens_user_name_unique UNIQUE(user_id, name)
);

-- Index for faster token validation queries
CREATE INDEX idx_api_tokens_active ON api_tokens (is_active) WHERE is_active = true;

-- Index for user token queries
CREATE INDEX idx_api_tokens_user_id ON api_tokens (user_id);