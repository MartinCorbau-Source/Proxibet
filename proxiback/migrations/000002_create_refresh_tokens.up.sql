CREATE TABLE IF NOT EXISTS refresh_tokens
(
    id UUID PRIMARY KEY,

    user_id UUID NOT NULL
        REFERENCES users (id)
        ON DELETE CASCADE,

    token_hash TEXT NOT NULL UNIQUE,

    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    replaced_by_token_id UUID
        REFERENCES refresh_tokens (id)
        ON DELETE SET NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS refresh_tokens_user_id_idx
    ON refresh_tokens (user_id);

CREATE INDEX IF NOT EXISTS refresh_tokens_active_user_id_idx
    ON refresh_tokens (user_id)
    WHERE revoked_at IS NULL;
