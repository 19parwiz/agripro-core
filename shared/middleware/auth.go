package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/19parwiz/agripro-core/shared/jwt"
	"github.com/19parwiz/agripro-core/shared/response"
)

type contextKey string

const claimsKey contextKey = "claims"

// Authenticate validates the Bearer token and stores claims on the request context.
// Use this on routes that require a logged-in user.
func Authenticate(jwtManager *jwt.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := bearerToken(r)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, err.Error())
				return
			}

			claims, err := jwtManager.Validate(token)
			if err != nil {
				message := "invalid token"
				if errors.Is(err, jwt.ErrExpiredToken) {
					message = "token expired"
				}
				response.Error(w, http.StatusUnauthorized, message)
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFromContext returns the JWT claims attached by Authenticate.
func ClaimsFromContext(ctx context.Context) (*jwt.Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*jwt.Claims)
	return claims, ok
}

// bearerToken extracts the token from an Authorization: Bearer <token> header.
func bearerToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", errors.New("missing authorization header")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header")
	}

	return parts[1], nil
}
