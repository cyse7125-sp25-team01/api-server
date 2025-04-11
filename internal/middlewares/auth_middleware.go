package middlewares

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/csye7125/team01/internal/store"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const UserContextKey contextKey = "user"

// Create a tracer for auth middleware
var tracer = otel.Tracer("github.com/csye7125/team01/internal/middlewares/auth")

type AuthMiddleware struct {
	UserStore *store.UserStore
}

func NewAuthMiddleware(userStore *store.UserStore) *AuthMiddleware {
	return &AuthMiddleware{UserStore: userStore}
}

func (am *AuthMiddleware) BasicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a span for authentication
		ctx, span := tracer.Start(r.Context(), "auth.basic", 
			trace.WithAttributes(
				attribute.String("auth.type", "basic"),
				attribute.String("request.path", r.URL.Path),
			),
		)
		defer span.End()

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			span.SetStatus(codes.Error, "missing authorization header")
			http.Error(w, `{"error": "Missing Authorization header"}`, http.StatusUnauthorized)
			return
		}

		// Decode Basic Auth
		authParts := strings.SplitN(authHeader, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Basic" {
			span.SetStatus(codes.Error, "invalid authorization header")
			http.Error(w, `{"error": "Invalid Authorization header"}`, http.StatusUnauthorized)
			return
		}

		payload, err := base64.StdEncoding.DecodeString(authParts[1])
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "invalid base64 encoding")
			http.Error(w, `{"error": "Invalid base64 encoding"}`, http.StatusUnauthorized)
			return
		}

		credentials := strings.SplitN(string(payload), ":", 2)
		if len(credentials) != 2 {
			span.SetStatus(codes.Error, "invalid credentials format")
			http.Error(w, `{"error": "Invalid credentials format"}`, http.StatusUnauthorized)
			return
		}

		username, password := credentials[0], credentials[1]
		span.SetAttributes(attribute.String("auth.username", username))

		// Check user in database
		user, err := am.UserStore.GetUserByUsername(ctx, username)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "invalid username")
			http.Error(w, `{"error": "Invalid username or password"}`, http.StatusUnauthorized)
			return
		}

		// Compare stored hashed password
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "invalid password")
			http.Error(w, `{"error": "Invalid username or password"}`, http.StatusUnauthorized)
			return
		}

		// Mark the span as successful
		span.SetStatus(codes.Ok, "authentication successful")
		span.SetAttributes(attribute.String("auth.user_id", fmt.Sprintf("%d", user.ID)))

		// Store user in context
		ctx = context.WithValue(ctx, UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}