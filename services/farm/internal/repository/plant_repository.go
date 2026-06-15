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

type PlantRepository interface {
	Create(ctx context.Context, plant *model.Plant) error
	ListByUser(ctx context.Context, userID string) ([]model.Plant, error)
	FindByID(ctx context.Context, userID, id string) (*model.Plant, error)
	Update(ctx context.Context, plant *model.Plant) error
	Delete(ctx context.Context, userID, id string) error
}

type plantRepository struct {
	pool *pgxpool.Pool
}

func NewPlantRepository(pool *pgxpool.Pool) PlantRepository {
	return &plantRepository{pool: pool}
}

func (r *plantRepository) Create(ctx context.Context, plant *model.Plant) error {
	if plant.ID == "" {
		plant.ID = uuid.NewString()
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO plants (
			id, user_id, name, variety, planting_date, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, plant.ID, plant.UserID, plant.Name, plant.Variety, plant.PlantingDate,
		plant.CreatedAt, plant.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert plant: %w", err)
	}

	return nil
}

func (r *plantRepository) ListByUser(ctx context.Context, userID string) ([]model.Plant, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, name, variety, planting_date, created_at, updated_at
		FROM plants
		WHERE user_id = $1
		ORDER BY planting_date DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list plants: %w", err)
	}
	defer rows.Close()

	var plants []model.Plant
	for rows.Next() {
		var p model.Plant
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Variety,
			&p.PlantingDate, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan plant: %w", err)
		}
		plants = append(plants, p)
	}

	return plants, rows.Err()
}

func (r *plantRepository) FindByID(ctx context.Context, userID, id string) (*model.Plant, error) {
	var p model.Plant

	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, name, variety, planting_date, created_at, updated_at
		FROM plants
		WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(&p.ID, &p.UserID, &p.Name, &p.Variety,
		&p.PlantingDate, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("find plant: %w", err)
	}

	return &p, nil
}

func (r *plantRepository) Update(ctx context.Context, plant *model.Plant) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE plants
		SET name = $1, variety = $2, planting_date = $3, updated_at = $4
		WHERE id = $5 AND user_id = $6
	`, plant.Name, plant.Variety, plant.PlantingDate, plant.UpdatedAt,
		plant.ID, plant.UserID)
	if err != nil {
		return fmt.Errorf("update plant: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}

	return nil
}

func (r *plantRepository) Delete(ctx context.Context, userID, id string) error {
	result, err := r.pool.Exec(ctx, `
		DELETE FROM plants WHERE id = $1 AND user_id = $2
	`, id, userID)
	if err != nil {
		return fmt.Errorf("delete plant: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}

	return nil
}
