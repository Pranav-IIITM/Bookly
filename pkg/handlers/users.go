package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"bookly-backend/pkg/middleware"
	"bookly-backend/pkg/models"

	"cloud.google.com/go/firestore"
	"github.com/go-chi/chi/v5"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UsersHandler groups endpoints that operate on the "users" Firestore
// collection.
type UsersHandler struct {
	Firestore *firestore.Client
}

// syncUserRequest is the JSON body accepted by Sync.
type syncUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Sync handles POST /api/users/sync.
// It upserts a user document keyed on the Firebase UID extracted from the
// auth token. If the user already exists, the existing document is returned
// unchanged. If the user is new, a new document is created.
func (h *UsersHandler) Sync(w http.ResponseWriter, r *http.Request) {
	firebaseUID, ok := middleware.FirebaseUID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	// Parse optional body — the frontend may send name/email overrides.
	var req syncUserRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	// Fall back to token claims when body fields are empty.
	if req.Name == "" {
		req.Name = middleware.Name(r.Context())
	}
	if req.Email == "" {
		req.Email = middleware.Email(r.Context())
	}

	// Check whether this Firebase user already has a Firestore document.
	users := h.Firestore.Collection("users")
	iter := users.Where("firebase_uid", "==", firebaseUID).Limit(1).Documents(r.Context())
	snapshot, err := iter.Next()
	iter.Stop()

	if err != iterator.Done {
		// A document was found — return it.
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not load user")
			return
		}

		var user models.User
		if err := snapshot.DataTo(&user); err != nil {
			writeError(w, http.StatusInternalServerError, "could not load user")
			return
		}
		user.ID = snapshot.Ref.ID
		writeJSON(w, http.StatusOK, map[string]any{"user": user})
		return
	}

	// No existing document — create a new one.
	userRef := users.NewDoc()
	user := models.User{
		ID:          userRef.ID,
		FirebaseUID: firebaseUID,
		Name:        req.Name,
		Email:       req.Email,
		CreatedAt:   time.Now().UTC(),
	}
	if _, err := userRef.Set(r.Context(), user); err != nil {
		writeError(w, http.StatusInternalServerError, "could not create user")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"user": user})
}

// GetByID handles GET /api/user/{id}.
// It fetches a single user document by its Firestore document ID.
func (h *UsersHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	docID := chi.URLParam(r, "id")
	if docID == "" {
		writeError(w, http.StatusBadRequest, "user id is required")
		return
	}

	snapshot, err := h.Firestore.Collection("users").Doc(docID).Get(r.Context())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not load user")
		return
	}

	var user models.User
	if err := snapshot.DataTo(&user); err != nil {
		writeError(w, http.StatusInternalServerError, "could not load user")
		return
	}
	user.ID = snapshot.Ref.ID

	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}
