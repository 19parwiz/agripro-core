-- Plants table for the farm service.

CREATE TABLE IF NOT EXISTS plants (
    id            UUID PRIMARY KEY,
    user_id       UUID         NOT NULL,
    name          VARCHAR(255) NOT NULL,
    variety       VARCHAR(255) NOT NULL DEFAULT '',
    planting_date DATE         NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_plants_user_id ON plants (user_id);
