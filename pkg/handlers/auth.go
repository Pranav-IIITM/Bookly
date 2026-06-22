package handlers

import (
	"net/http"

	"bookly-backend/pkg/middleware"
)

// AuthHandler groups endpoints related to authentication verification.
type AuthHandler struct{}

// Verify handles POST /api/auth/verify.
// It is a protected endpoint — the auth middleware has already validated the
// token by the time we get here. We simply read the decoded claims from the
// request context and return them to the caller.
func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.FirebaseUID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"uid":   uid,
		"email": middleware.Email(r.Context()),
		"name":  middleware.Name(r.Context()),
	})
}
