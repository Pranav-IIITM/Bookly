package handlers

import (
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DataHandler provides generic CRUD operations against any Firestore
// collection. The target collection is specified via the "collection"
// query parameter (e.g. ?collection=products).
//
// All endpoints are protected by the auth middleware.
type DataHandler struct {
	Firestore *firestore.Client
}

// collection extracts and validates the "collection" query parameter.
func collection(r *http.Request) (string, bool) {
	c := r.URL.Query().Get("collection")
	return c, c != ""
}

// Create handles POST /api/data?collection=<name>.
// It creates a new document in the specified collection from the JSON body.
func (h *DataHandler) Create(w http.ResponseWriter, r *http.Request) {
	col, ok := collection(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "query parameter 'collection' is required")
		return
	}

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	ref := h.Firestore.Collection(col).NewDoc()
	if _, err := ref.Set(r.Context(), body); err != nil {
		writeError(w, http.StatusInternalServerError, "could not create document")
		return
	}

	body["id"] = ref.ID
	writeJSON(w, http.StatusCreated, map[string]any{"document": body})
}

// Read handles GET /api/data/{id}?collection=<name>.
// It fetches a single document by ID from the specified collection.
func (h *DataHandler) Read(w http.ResponseWriter, r *http.Request) {
	col, ok := collection(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "query parameter 'collection' is required")
		return
	}

	docID := chi.URLParam(r, "id")
	if docID == "" {
		writeError(w, http.StatusBadRequest, "document id is required")
		return
	}

	snapshot, err := h.Firestore.Collection(col).Doc(docID).Get(r.Context())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			writeError(w, http.StatusNotFound, "document not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not read document")
		return
	}

	data := snapshot.Data()
	data["id"] = snapshot.Ref.ID
	writeJSON(w, http.StatusOK, map[string]any{"document": data})
}

// Update handles PUT /api/data/{id}?collection=<name>.
// It merges the JSON body into the existing document (MergeAll).
func (h *DataHandler) Update(w http.ResponseWriter, r *http.Request) {
	col, ok := collection(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "query parameter 'collection' is required")
		return
	}

	docID := chi.URLParam(r, "id")
	if docID == "" {
		writeError(w, http.StatusBadRequest, "document id is required")
		return
	}

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	ref := h.Firestore.Collection(col).Doc(docID)

	// Verify the document exists before updating.
	if _, err := ref.Get(r.Context()); err != nil {
		if status.Code(err) == codes.NotFound {
			writeError(w, http.StatusNotFound, "document not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not update document")
		return
	}

	if _, err := ref.Set(r.Context(), body, firestore.MergeAll); err != nil {
		writeError(w, http.StatusInternalServerError, "could not update document")
		return
	}

	body["id"] = docID
	writeJSON(w, http.StatusOK, map[string]any{"document": body})
}

// Delete handles DELETE /api/data/{id}?collection=<name>.
// It removes the document from the specified collection.
func (h *DataHandler) Delete(w http.ResponseWriter, r *http.Request) {
	col, ok := collection(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "query parameter 'collection' is required")
		return
	}

	docID := chi.URLParam(r, "id")
	if docID == "" {
		writeError(w, http.StatusBadRequest, "document id is required")
		return
	}

	ref := h.Firestore.Collection(col).Doc(docID)

	// Verify the document exists before deleting.
	if _, err := ref.Get(r.Context()); err != nil {
		if status.Code(err) == codes.NotFound {
			writeError(w, http.StatusNotFound, "document not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not delete document")
		return
	}

	if _, err := ref.Delete(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "could not delete document")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "document deleted"})
}
