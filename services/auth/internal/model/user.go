package model

import "time"

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// User represents a registered AgriPro account stored in PostgreSQL.
type User struct {
	ID                string
	Email             string
	PasswordHash      string
	FullName          string
	Role              string
	EmailVerified     bool
	VerificationToken string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// NewUser creates a user with sensible defaults for registration.
func NewUser(email, passwordHash, fullName string) User {
	now := time.Now().UTC()

	return User{
		Email:         email,
		PasswordHash:  passwordHash,
		FullName:      fullName,
		Role:          RoleUser,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}
