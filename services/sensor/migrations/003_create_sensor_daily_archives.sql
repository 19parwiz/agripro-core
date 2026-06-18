-- Daily aggregated sensor stats for history charts.

CREATE TABLE IF NOT EXISTS sensor_daily_archives (
    id            UUID PRIMARY KEY,
    device_id     UUID             NOT NULL,
    sensor_type   VARCHAR(50)      NOT NULL,
    bucket_date   DATE             NOT NULL,
    avg_value     DOUBLE PRECISION NOT NULL,
    min_value     DOUBLE PRECISION NOT NULL,
    max_value     DOUBLE PRECISION NOT NULL,
    reading_count INTEGER          NOT NULL,
    created_at    TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    UNIQUE (device_id, sensor_type, bucket_date)
);

CREATE INDEX IF NOT EXISTS idx_sensor_daily_device_type_date
    ON sensor_daily_archives (device_id, sensor_type, bucket_date);
