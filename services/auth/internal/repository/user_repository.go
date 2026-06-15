package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/19parwiz/agripro-core/services/auth/internal/model"
	apperrors "github.com/19parwiz/agripro-core/shared/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository defines database operations for users.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByVerificationToken(ctx context.Context, token string) (*model.User, error)
	MarkEmailVerified(ctx context.Context, userID string) error
	EmailExists(ctx context.Context, email string) (bool, error)
}

type userRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a PostgreSQL-backed user repository.
func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	if user.ID == "" {
		user.ID = uuid.NewString()
	}

	query := `
		INSERT INTO users (
			id, email, password_hash, full_name, role,
			email_verified, verification_token, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Role,
		user.EmailVerified,
		user.VerificationToken,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.findOne(ctx, `
		SELECT id, email, password_hash, full_name, role,
		       email_verified, verification_token, created_at, updated_at
		FROM users
		WHERE email = $1
	`, email)
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	return r.findOne(ctx, `
		SELECT id, email, password_hash, full_name, role,
		       email_verified, verification_token, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id)
}

func (r *userRepository) FindByVerificationToken(ctx context.Context, token string) (*model.User, error) {
	return r.findOne(ctx, `
		SELECT id, email, password_hash, full_name, role,
		       email_verified, verification_token, created_at, updated_at
		FROM users
		WHERE verification_token = $1
	`, token)
}

func (r *userRepository) MarkEmailVerified(ctx context.Context, userID string) error {
	query := `
		UPDATE users
		SET email_verified = TRUE,
		    verification_token = NULL,
		    updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("verify user email: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}

	return nil
}

func (r *userRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool

	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)
	`, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check email exists: %w", err)
	}

	return exists, nil
}

func (r *userRepository) findOne(ctx context.Context, query string, arg any) (*model.User, error) {
	var user model.User

	err := r.pool.QueryRow(ctx, query, arg).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.EmailVerified,
		&user.VerificationToken,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("query user: %w", err)
	}

	return &user, nil
}
