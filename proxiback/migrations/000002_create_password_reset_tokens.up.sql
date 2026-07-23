CREATE TABLE password_reset_tokens
(
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash CHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX password_reset_tokens_user_id_idx
    ON password_reset_tokens (user_id);

CREATE INDEX password_reset_tokens_valid_idx
    ON password_reset_tokens (token_hash)
    WHERE used_at IS NULL;