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

type ReadingRepository interface {
	Create(ctx context.Context, reading *model.Reading) error
	CreateBatch(ctx context.Context, readings []model.Reading) error
	ListByDeviceAndType(ctx context.Context, deviceID, sensorType string, from, to time.Time) ([]model.Reading, error)
	AggregateForHour(ctx context.Context, deviceID, sensorType string, bucketStart time.Time) (*model.HourlyArchive, error)
	AggregateForDay(ctx context.Context, deviceID, sensorType string, bucketDate time.Time) (*model.DailyArchive, error)
}

type readingRepository struct {
	pool *pgxpool.Pool
}

func NewReadingRepository(pool *pgxpool.Pool) ReadingRepository {
	return &readingRepository{pool: pool}
}

func (r *readingRepository) Create(ctx context.Context, reading *model.Reading) error {
	if reading.ID == "" {
		reading.ID = uuid.NewString()
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO sensor_readings (
			id, device_id, sensor_type, value, unit, recorded_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, reading.ID, reading.DeviceID, reading.SensorType, reading.Value, reading.Unit,
		reading.RecordedAt, reading.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert sensor reading: %w", err)
	}

	return nil
}

func (r *readingRepository) CreateBatch(ctx context.Context, readings []model.Reading) error {
	if len(readings) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin batch insert: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	for i := range readings {
		if readings[i].ID == "" {
			readings[i].ID = uuid.NewString()
		}

		batch.Queue(`
			INSERT INTO sensor_readings (
				id, device_id, sensor_type, value, unit, recorded_at, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, readings[i].ID, readings[i].DeviceID, readings[i].SensorType, readings[i].Value,
			readings[i].Unit, readings[i].RecordedAt, readings[i].CreatedAt)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for range readings {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("insert sensor reading batch: %w", err)
		}
	}

	if err := br.Close(); err != nil {
		return fmt.Errorf("close sensor reading batch: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit sensor reading batch: %w", err)
	}

	return nil
}

func (r *readingRepository) ListByDeviceAndType(
	ctx context.Context,
	deviceID, sensorType string,
	from, to time.Time,
) ([]model.Reading, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, device_id, sensor_type, value, unit, recorded_at, created_at
		FROM sensor_readings
		WHERE device_id = $1
		  AND sensor_type = $2
		  AND recorded_at >= $3
		  AND recorded_at < $4
		ORDER BY recorded_at ASC
	`, deviceID, sensorType, from, to)
	if err != nil {
		return nil, fmt.Errorf("list sensor readings: %w", err)
	}
	defer rows.Close()

	return scanReadings(rows)
}

func (r *readingRepository) AggregateForHour(
	ctx context.Context,
	deviceID, sensorType string,
	bucketStart time.Time,
) (*model.HourlyArchive, error) {
	bucketEnd := bucketStart.Add(time.Hour)

	var count int
	var avg, min, max float64
	err := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(AVG(value), 0),
			COALESCE(MIN(value), 0),
			COALESCE(MAX(value), 0),
			COUNT(*)::int
		FROM sensor_readings
		WHERE device_id = $1
		  AND sensor_type = $2
		  AND recorded_at >= $3
		  AND recorded_at < $4
	`, deviceID, sensorType, bucketStart, bucketEnd).Scan(&avg, &min, &max, &count)
	if err != nil {
		return nil, fmt.Errorf("aggregate hourly readings: %w", err)
	}

	if count == 0 {
		return nil, nil
	}

	return &model.HourlyArchive{
		ID:           uuid.NewString(),
		DeviceID:     deviceID,
		SensorType:   sensorType,
		BucketStart:  bucketStart,
		AvgValue:     avg,
		MinValue:     min,
		MaxValue:     max,
		ReadingCount: count,
	}, nil
}

func (r *readingRepository) AggregateForDay(
	ctx context.Context,
	deviceID, sensorType string,
	bucketDate time.Time,
) (*model.DailyArchive, error) {
	dayStart := time.Date(bucketDate.Year(), bucketDate.Month(), bucketDate.Day(), 0, 0, 0, 0, bucketDate.Location())
	dayEnd := dayStart.Add(24 * time.Hour)

	var count int
	var avg, min, max float64
	err := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(AVG(value), 0),
			COALESCE(MIN(value), 0),
			COALESCE(MAX(value), 0),
			COUNT(*)::int
		FROM sensor_readings
		WHERE device_id = $1
		  AND sensor_type = $2
		  AND recorded_at >= $3
		  AND recorded_at < $4
	`, deviceID, sensorType, dayStart, dayEnd).Scan(&avg, &min, &max, &count)
	if err != nil {
		return nil, fmt.Errorf("aggregate daily readings: %w", err)
	}

	if count == 0 {
		return nil, nil
	}

	return &model.DailyArchive{
		ID:           uuid.NewString(),
		DeviceID:     deviceID,
		SensorType:   sensorType,
		BucketDate:   dayStart,
		AvgValue:     avg,
		MinValue:     min,
		MaxValue:     max,
		ReadingCount: count,
	}, nil
}

func scanReadings(rows pgx.Rows) ([]model.Reading, error) {
	var readings []model.Reading
	for rows.Next() {
		var reading model.Reading
		if err := rows.Scan(
			&reading.ID,
			&reading.DeviceID,
			&reading.SensorType,
			&reading.Value,
			&reading.Unit,
			&reading.RecordedAt,
			&reading.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan sensor reading: %w", err)
		}
		readings = append(readings, reading)
	}

	return readings, rows.Err()
}
