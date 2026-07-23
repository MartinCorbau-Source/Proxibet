CREATE TABLE IF NOT EXISTS groups
(
    id UUID PRIMARY KEY,

    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',

    owner_id UUID NOT NULL
        REFERENCES users (id)
        ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS groups_owner_id_idx
    ON groups (owner_id);

CREATE TABLE IF NOT EXISTS group_members
(
    group_id UUID NOT NULL
        REFERENCES groups (id)
        ON DELETE CASCADE,

    user_id UUID NOT NULL
        REFERENCES users (id)
        ON DELETE CASCADE,

    role VARCHAR(20) NOT NULL,

    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (group_id, user_id),

    CONSTRAINT group_members_role_check
        CHECK (role IN ('OWNER', 'ADMIN', 'MEMBER'))
);

CREATE INDEX IF NOT EXISTS group_members_user_id_idx
    ON group_members (user_id);

CREATE INDEX IF NOT EXISTS group_members_group_id_role_idx
    ON group_members (group_id, role);

CREATE UNIQUE INDEX IF NOT EXISTS group_members_single_owner_idx
    ON group_members (group_id)
    WHERE role = 'OWNER';

CREATE TABLE IF NOT EXISTS group_invitations
(
    id UUID PRIMARY KEY,

    group_id UUID NOT NULL
        REFERENCES groups (id)
        ON DELETE CASCADE,

    code VARCHAR(64) NOT NULL,

    expires_at TIMESTAMPTZ NOT NULL,
    max_uses INTEGER NOT NULL,
    uses_count INTEGER NOT NULL DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT TRUE,

    created_by UUID NOT NULL
        REFERENCES users (id)
        ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deactivated_at TIMESTAMPTZ,

    CONSTRAINT group_invitations_max_uses_check
        CHECK (max_uses > 0),

    CONSTRAINT group_invitations_uses_count_check
        CHECK (uses_count >= 0 AND uses_count <= max_uses)
);

CREATE UNIQUE INDEX IF NOT EXISTS group_invitations_code_unique
    ON group_invitations (code);

CREATE INDEX IF NOT EXISTS group_invitations_group_id_idx
    ON group_invitations (group_id);

CREATE INDEX IF NOT EXISTS group_invitations_active_code_idx
    ON group_invitations (code)
    WHERE active = TRUE;


