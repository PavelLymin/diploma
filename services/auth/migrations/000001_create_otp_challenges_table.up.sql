CREATE TABLE auth_otps (
    id UUID PRIMARY KEY,
    phone VARCHAR(16) NOT NULL,
    code_hash BYTEA NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    consumed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_otps_phone_created
    ON auth_otps (phone, created_at DESC);