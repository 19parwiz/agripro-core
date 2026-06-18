-- Hourly aggregated sensor stats for history charts.

CREATE TABLE IF NOT EXISTS sensor_hourly_archives (
    id            UUID PRIMARY KEY,
    device_id     UUID             NOT NULL,
    sensor_type   VARCHAR(50)      NOT NULL,
    bucket_start  TIMESTAMPTZ      NOT NULL,
    avg_value     DOUBLE PRECISION NOT NULL,
    min_value     DOUBLE PRECISION NOT NULL,
    max_value     DOUBLE PRECISION NOT NULL,
    reading_count INTEGER          NOT NULL,
    created_at    TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    UNIQUE (device_id, sensor_type, bucket_start)
);

CREATE INDEX IF NOT EXISTS idx_sensor_hourly_device_type_time
    ON sensor_hourly_archives (device_id, sensor_type, bucket_start);
