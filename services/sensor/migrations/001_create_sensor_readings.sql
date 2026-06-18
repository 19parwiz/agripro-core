-- Raw sensor readings ingested from the VPS snapshot.

CREATE TABLE IF NOT EXISTS sensor_readings (
    id          UUID PRIMARY KEY,
    device_id   UUID             NOT NULL,
    sensor_type VARCHAR(50)      NOT NULL,
    value       DOUBLE PRECISION NOT NULL,
    unit        VARCHAR(20)      NOT NULL DEFAULT '',
    recorded_at TIMESTAMPTZ      NOT NULL,
    created_at  TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sensor_readings_device_type_time
    ON sensor_readings (device_id, sensor_type, recorded_at);
