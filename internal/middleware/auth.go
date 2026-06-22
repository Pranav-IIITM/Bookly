// Package middleware provides HTTP middleware for the Bookly API.
// The primary export is FirebaseAuth, which verifies Firebase ID tokens
// and attaches the decoded claims to the request context.
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
)

// contextKey is an unexported type to prevent collisions with context keys
// defined in other packages.
type contextKey string

const (
	firebaseUIDKey contextKey = "firebaseUID"
	emailKey       contextKey = "email"
	nameKey        contextKey = "name"
)

// FirebaseAuth returns middleware that validates the Authorization header
// against Firebase Authentication. Requests without a valid Bearer token
// are rejected with 401 Unauthorized.
func FirebaseAuth(authClient *auth.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// ── Extract the Bearer token ────────────────────────────
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.Fields(authHeader)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeError(w, http.StatusUnauthorized, "authorization header must be Bearer token")
				return
			}

			// ── Verify the token with Firebase ──────────────────────
			token, err := authClient.VerifyIDToken(r.Context(), parts[1])
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			// ── Attach decoded claims to request context ────────────
			ctx := context.WithValue(r.Context(), firebaseUIDKey, token.UID)
			if email, ok := token.Claims["email"].(string); ok {
				ctx = context.WithValue(ctx, emailKey, email)
			}
			if name, ok := token.Claims["name"].(string); ok {
				ctx = context.WithValue(ctx, nameKey, name)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// --------------------------------------------------------------------------
// Context value accessors
// --------------------------------------------------------------------------

// FirebaseUID extracts the authenticated user's Firebase UID from the
// request context. The second return value is false when no UID is present.
func FirebaseUID(ctx context.Context) (string, bool) {
	uid, ok := ctx.Value(firebaseUIDKey).(string)
	return uid, ok && uid != ""
}

// Email extracts the user's email from the request context.
func Email(ctx context.Context) string {
	email, _ := ctx.Value(emailKey).(string)
	return email
}

// Name extracts the user's display name from the request context.
func Name(ctx context.Context) string {
	name, _ := ctx.Value(nameKey).(string)
	return name
}

// --------------------------------------------------------------------------
// Internal helpers
// --------------------------------------------------------------------------

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
