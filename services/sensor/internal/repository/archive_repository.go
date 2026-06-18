package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/19parwiz/agripro-core/services/sensor/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ArchiveRepository interface {
	UpsertHourly(ctx context.Context, archive *model.HourlyArchive) error
	ListHourly(ctx context.Context, deviceID, sensorType string, from, to time.Time) ([]model.HourlyArchive, error)
	UpsertDaily(ctx context.Context, archive *model.DailyArchive) error
	ListDaily(ctx context.Context, deviceID, sensorType string, from, to time.Time) ([]model.DailyArchive, error)
}

type archiveRepository struct {
	pool *pgxpool.Pool
}

func NewArchiveRepository(pool *pgxpool.Pool) ArchiveRepository {
	return &archiveRepository{pool: pool}
}

func (r *archiveRepository) UpsertHourly(ctx context.Context, archive *model.HourlyArchive) error {
	if archive.ID == "" {
		archive.ID = uuid.NewString()
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO sensor_hourly_archives (
			id, device_id, sensor_type, bucket_start,
			avg_value, min_value, max_value, reading_count, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (device_id, sensor_type, bucket_start)
		DO UPDATE SET
			avg_value = EXCLUDED.avg_value,
			min_value = EXCLUDED.min_value,
			max_value = EXCLUDED.max_value,
			reading_count = EXCLUDED.reading_count
	`, archive.ID, archive.DeviceID, archive.SensorType, archive.BucketStart,
		archive.AvgValue, archive.MinValue, archive.MaxValue, archive.ReadingCount, archive.CreatedAt)
	if err != nil {
		return fmt.Errorf("upsert hourly archive: %w", err)
	}

	return nil
}

func (r *archiveRepository) ListHourly(
	ctx context.Context,
	deviceID, sensorType string,
	from, to time.Time,
) ([]model.HourlyArchive, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, device_id, sensor_type, bucket_start,
		       avg_value, min_value, max_value, reading_count, created_at
		FROM sensor_hourly_archives
		WHERE device_id = $1
		  AND sensor_type = $2
		  AND bucket_start >= $3
		  AND bucket_start < $4
		ORDER BY bucket_start ASC
	`, deviceID, sensorType, from, to)
	if err != nil {
		return nil, fmt.Errorf("list hourly archives: %w", err)
	}
	defer rows.Close()

	return scanHourlyArchives(rows)
}

func (r *archiveRepository) UpsertDaily(ctx context.Context, archive *model.DailyArchive) error {
	if archive.ID == "" {
		archive.ID = uuid.NewString()
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO sensor_daily_archives (
			id, device_id, sensor_type, bucket_date,
			avg_value, min_value, max_value, reading_count, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (device_id, sensor_type, bucket_date)
		DO UPDATE SET
			avg_value = EXCLUDED.avg_value,
			min_value = EXCLUDED.min_value,
			max_value = EXCLUDED.max_value,
			reading_count = EXCLUDED.reading_count
	`, archive.ID, archive.DeviceID, archive.SensorType, archive.BucketDate,
		archive.AvgValue, archive.MinValue, archive.MaxValue, archive.ReadingCount, archive.CreatedAt)
	if err != nil {
		return fmt.Errorf("upsert daily archive: %w", err)
	}

	return nil
}

func (r *archiveRepository) ListDaily(
	ctx context.Context,
	deviceID, sensorType string,
	from, to time.Time,
) ([]model.DailyArchive, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, device_id, sensor_type, bucket_date,
		       avg_value, min_value, max_value, reading_count, created_at
		FROM sensor_daily_archives
		WHERE device_id = $1
		  AND sensor_type = $2
		  AND bucket_date >= $3::date
		  AND bucket_date < $4::date
		ORDER BY bucket_date ASC
	`, deviceID, sensorType, from, to)
	if err != nil {
		return nil, fmt.Errorf("list daily archives: %w", err)
	}
	defer rows.Close()

	return scanDailyArchives(rows)
}

func scanHourlyArchives(rows pgx.Rows) ([]model.HourlyArchive, error) {
	var archives []model.HourlyArchive
	for rows.Next() {
		var archive model.HourlyArchive
		if err := rows.Scan(
			&archive.ID,
			&archive.DeviceID,
			&archive.SensorType,
			&archive.BucketStart,
			&archive.AvgValue,
			&archive.MinValue,
			&archive.MaxValue,
			&archive.ReadingCount,
			&archive.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan hourly archive: %w", err)
		}
		archives = append(archives, archive)
	}

	return archives, rows.Err()
}

func scanDailyArchives(rows pgx.Rows) ([]model.DailyArchive, error) {
	var archives []model.DailyArchive
	for rows.Next() {
		var archive model.DailyArchive
		if err := rows.Scan(
			&archive.ID,
			&archive.DeviceID,
			&archive.SensorType,
			&archive.BucketDate,
			&archive.AvgValue,
			&archive.MinValue,
			&archive.MaxValue,
			&archive.ReadingCount,
			&archive.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan daily archive: %w", err)
		}
		archives = append(archives, archive)
	}

	return archives, rows.Err()
}
