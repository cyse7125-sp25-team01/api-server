package middlewares

import (
	"context"
	"encoding/base64"
	"github.com/csye7125/team01/internal/store"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
)

type contextKey string

const UserContextKey contextKey = "user"

type AuthMiddleware struct {
	UserStore *store.UserStore
}

func NewAuthMiddleware(userStore *store.UserStore) *AuthMiddleware {
	return &AuthMiddleware{UserStore: userStore}
}

func (am *AuthMiddleware) BasicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Missing Authorization header"}`, http.StatusUnauthorized)
			return
		}

		// Decode Basic Auth
		authParts := strings.SplitN(authHeader, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Basic" {
			http.Error(w, `{"error": "Invalid Authorization header"}`, http.StatusUnauthorized)
			return
		}

		payload, err := base64.StdEncoding.DecodeString(authParts[1])
		if err != nil {
			http.Error(w, `{"error": "Invalid base64 encoding"}`, http.StatusUnauthorized)
			return
		}

		credentials := strings.SplitN(string(payload), ":", 2)
		if len(credentials) != 2 {
			http.Error(w, `{"error": "Invalid credentials format"}`, http.StatusUnauthorized)
			return
		}

		username, password := credentials[0], credentials[1]

		// Check user in database
		user, err := am.UserStore.GetUserByUsername(r.Context(), username)
		if err != nil {
			http.Error(w, `{"error": "Invalid username or password"}`, http.StatusUnauthorized)
			return
		}

		// Compare stored hashed password
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
			http.Error(w, `{"error": "Invalid username or password"}`, http.StatusUnauthorized)
			return
		}

		// Store user in context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
