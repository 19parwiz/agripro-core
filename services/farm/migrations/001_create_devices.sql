-- Devices table for the farm service.

CREATE TABLE IF NOT EXISTS devices (
    id          UUID PRIMARY KEY,
    user_id     UUID         NOT NULL,
    name        VARCHAR(255) NOT NULL,
    device_id   VARCHAR(255) NOT NULL,
    type        VARCHAR(100) NOT NULL DEFAULT '',
    location    VARCHAR(255) NOT NULL DEFAULT '',
    stream_path VARCHAR(255) NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_devices_user_id ON devices (user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_devices_user_device_id ON devices (user_id, device_id);
