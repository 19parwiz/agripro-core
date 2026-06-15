package service

import (
	"context"
	"net/http"
	"strings"

	"github.com/19parwiz/agripro-core/services/auth/internal/model"
	"github.com/19parwiz/agripro-core/services/auth/internal/repository"
	apperrors "github.com/19parwiz/agripro-core/shared/errors"
	"github.com/19parwiz/agripro-core/shared/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const minPasswordLength = 8

// AuthService handles registration, login, and email verification.
type AuthService struct {
	users      repository.UserRepository
	jwtManager *jwt.Manager
}

func NewAuthService(users repository.UserRepository, jwtManager *jwt.Manager) *AuthService {
	return &AuthService{
		users:      users,
		jwtManager: jwtManager,
	}
}

type RegisterInput struct {
	Email    string
	Password string
	FullName string
}

type LoginInput struct {
	Email    string
	Password string
}

// UserProfile is the safe user data returned to clients.
type UserProfile struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	FullName      string `json:"full_name"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
}

type RegisterResult struct {
	User UserProfile `json:"user"`
}

type LoginResult struct {
	Token string      `json:"token"`
	User  UserProfile `json:"user"`
}

// Register creates a new account and stores a verification token.
func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*RegisterResult, error) {
	email := normalizeEmail(input.Email)
	if err := validateCredentials(email, input.Password); err != nil {
		return nil, err
	}

	exists, err := s.users.EmailExists(ctx, email)
	if err != nil {
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "could not check email")
	}
	if exists {
		return nil, apperrors.New(http.StatusConflict, "email already registered")
	}

	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "could not hash password")
	}

	user := model.NewUser(email, passwordHash, strings.TrimSpace(input.FullName))
	user.VerificationToken = uuid.NewString()

	if err := s.users.Create(ctx, &user); err != nil {
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "could not create user")
	}

	return &RegisterResult{User: toProfile(user)}, nil
}

// Login checks credentials and returns a JWT for verified users.
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	email := normalizeEmail(input.Email)
	if email == "" || input.Password == "" {
		return nil, apperrors.ErrBadRequest
	}

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	if !user.EmailVerified {
		return nil, apperrors.ErrEmailNotVerified
	}

	token, err := s.jwtManager.Generate(user.ID, user.Email)
	if err != nil {
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "could not create token")
	}

	return &LoginResult{
		Token: token,
		User:  toProfile(*user),
	}, nil
}

// VerifyEmail marks a user as verified using the token from their email.
func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return apperrors.ErrBadRequest
	}

	user, err := s.users.FindByVerificationToken(ctx, token)
	if err != nil {
		return err
	}

	if user.EmailVerified {
		return apperrors.New(http.StatusConflict, "email already verified")
	}

	return s.users.MarkEmailVerified(ctx, user.ID)
}

func validateCredentials(email, password string) error {
	if email == "" || password == "" {
		return apperrors.ErrBadRequest
	}
	if len(password) < minPasswordLength {
		return apperrors.New(http.StatusBadRequest, "password must be at least 8 characters")
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func toProfile(user model.User) UserProfile {
	return UserProfile{
		ID:            user.ID,
		Email:         user.Email,
		FullName:      user.FullName,
		Role:          user.Role,
		EmailVerified: user.EmailVerified,
	}
}
