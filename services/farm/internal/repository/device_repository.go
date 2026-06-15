package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/19parwiz/agripro-core/services/farm/internal/model"
	apperrors "github.com/19parwiz/agripro-core/shared/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeviceRepository interface {
	Create(ctx context.Context, device *model.Device) error
	ListByUser(ctx context.Context, userID string) ([]model.Device, error)
	FindByID(ctx context.Context, userID, id string) (*model.Device, error)
	Update(ctx context.Context, device *model.Device) error
	Delete(ctx context.Context, userID, id string) error
}

type deviceRepository struct {
	pool *pgxpool.Pool
}

func NewDeviceRepository(pool *pgxpool.Pool) DeviceRepository {
	return &deviceRepository{pool: pool}
}

func (r *deviceRepository) Create(ctx context.Context, device *model.Device) error {
	if device.ID == "" {
		device.ID = uuid.NewString()
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO devices (
			id, user_id, name, device_id, type, location, stream_path, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, device.ID, device.UserID, device.Name, device.DeviceID, device.Type,
		device.Location, device.StreamPath, device.CreatedAt, device.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert device: %w", err)
	}

	return nil
}

func (r *deviceRepository) ListByUser(ctx context.Context, userID string) ([]model.Device, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, name, device_id, type, location, stream_path, created_at, updated_at
		FROM devices
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	var devices []model.Device
	for rows.Next() {
		var d model.Device
		if err := rows.Scan(&d.ID, &d.UserID, &d.Name, &d.DeviceID, &d.Type,
			&d.Location, &d.StreamPath, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan device: %w", err)
		}
		devices = append(devices, d)
	}

	return devices, rows.Err()
}

func (r *deviceRepository) FindByID(ctx context.Context, userID, id string) (*model.Device, error) {
	var d model.Device

	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, name, device_id, type, location, stream_path, created_at, updated_at
		FROM devices
		WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(&d.ID, &d.UserID, &d.Name, &d.DeviceID, &d.Type,
		&d.Location, &d.StreamPath, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("find device: %w", err)
	}

	return &d, nil
}

func (r *deviceRepository) Update(ctx context.Context, device *model.Device) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE devices
		SET name = $1, device_id = $2, type = $3, location = $4,
		    stream_path = $5, updated_at = $6
		WHERE id = $7 AND user_id = $8
	`, device.Name, device.DeviceID, device.Type, device.Location,
		device.StreamPath, device.UpdatedAt, device.ID, device.UserID)
	if err != nil {
		return fmt.Errorf("update device: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}

	return nil
}

func (r *deviceRepository) Delete(ctx context.Context, userID, id string) error {
	result, err := r.pool.Exec(ctx, `
		DELETE FROM devices WHERE id = $1 AND user_id = $2
	`, id, userID)
	if err != nil {
		return fmt.Errorf("delete device: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}

	return nil
}
