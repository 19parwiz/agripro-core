-- Users table for the auth service.

CREATE TABLE IF NOT EXISTS users (
    id                 UUID PRIMARY KEY,
    email              VARCHAR(255) NOT NULL UNIQUE,
    password_hash      VARCHAR(255) NOT NULL,
    full_name          VARCHAR(255) NOT NULL DEFAULT '',
    role               VARCHAR(50)  NOT NULL DEFAULT 'user',
    email_verified     BOOLEAN      NOT NULL DEFAULT FALSE,
    verification_token VARCHAR(255),
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_verification_token ON users (verification_token);
