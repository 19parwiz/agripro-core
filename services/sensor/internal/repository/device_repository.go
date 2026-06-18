package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DeviceRepository reads device rows owned by the farm service.
type DeviceRepository interface {
	ListIDs(ctx context.Context) ([]string, error)
	BelongsToUser(ctx context.Context, deviceID, userID string) (bool, error)
}

type deviceRepository struct {
	pool *pgxpool.Pool
}

func NewDeviceRepository(pool *pgxpool.Pool) DeviceRepository {
	return &deviceRepository{pool: pool}
}

func (r *deviceRepository) ListIDs(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT id FROM devices ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("list device ids: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan device id: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

func (r *deviceRepository) BelongsToUser(ctx context.Context, deviceID, userID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM devices WHERE id = $1 AND user_id = $2
		)
	`, deviceID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check device ownership: %w", err)
	}

	return exists, nil
}
