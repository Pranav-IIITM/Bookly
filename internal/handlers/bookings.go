package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"bookly-backend/internal/middleware"
	"bookly-backend/internal/models"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BookingsHandler groups endpoints that operate on the "bookings" Firestore
// collection.
type BookingsHandler struct {
	Firestore *firestore.Client
}

// createBookingRequest accepts the slot ID as either a JSON string or number
// to be resilient to different frontend serialisation styles.
type createBookingRequest struct {
	SlotID json.RawMessage `json:"slotId"`
}

// Sentinel errors used inside the Firestore transaction.
var (
	errSlotFull     = errors.New("slot full")
	errUserNotFound = errors.New("user not found")
)

// Create handles POST /api/book.
// It atomically verifies slot capacity, creates the booking document, and
// increments the slot's booked count — all inside a Firestore transaction.
func (h *BookingsHandler) Create(w http.ResponseWriter, r *http.Request) {
	firebaseUID, ok := middleware.FirebaseUID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var req createBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	slotID, err := parseIDField(req.SlotID, "slotId")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// ── Firestore transaction ───────────────────────────────────────────
	var booking models.Booking
	err = h.Firestore.RunTransaction(r.Context(), func(ctx context.Context, tx *firestore.Transaction) error {
		// 1. Look up the Firestore user by Firebase UID.
		userQuery := h.Firestore.Collection("users").Where("firebase_uid", "==", firebaseUID).Limit(1)
		userSnapshots, err := tx.Documents(userQuery).GetAll()
		if err != nil {
			return err
		}
		if len(userSnapshots) == 0 {
			return errUserNotFound
		}
		userSnapshot := userSnapshots[0]
		var user models.User
		if err := userSnapshot.DataTo(&user); err != nil {
			return err
		}
		user.ID = userSnapshot.Ref.ID

		// 2. Load the slot and verify capacity.
		slotRef := h.Firestore.Collection("slots").Doc(slotID)
		slotSnapshot, err := tx.Get(slotRef)
		if err != nil {
			return err
		}
		var slot models.Slot
		if err := slotSnapshot.DataTo(&slot); err != nil {
			return err
		}
		slot.ID = slotSnapshot.Ref.ID
		if slot.BookedCount >= slot.Capacity {
			return errSlotFull
		}

		// 3. Create the booking document.
		bookingRef := h.Firestore.Collection("bookings").NewDoc()
		booking = models.Booking{
			ID:        bookingRef.ID,
			UserID:    user.ID,
			SlotID:    slot.ID,
			Status:    "confirmed",
			CreatedAt: time.Now().UTC(),
			Slot:      &slot,
		}
		if err := tx.Set(bookingRef, booking); err != nil {
			return err
		}

		// 4. Atomically increment the slot's booked count.
		return tx.Update(slotRef, []firestore.Update{
			{Path: "bookedCount", Value: firestore.Increment(1)},
		})
	})

	if err != nil {
		switch {
		case status.Code(err) == codes.NotFound, errors.Is(err, errUserNotFound):
			writeError(w, http.StatusNotFound, "user or slot not found")
		case errors.Is(err, errSlotFull):
			writeError(w, http.StatusBadRequest, "slot is already full")
		default:
			writeError(w, http.StatusInternalServerError, "could not create booking")
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"booking": booking})
}

// List handles GET /api/bookings.
// It returns all bookings for the authenticated user, enriched with the
// related slot data, sorted newest-first.
func (h *BookingsHandler) List(w http.ResponseWriter, r *http.Request) {
	firebaseUID, ok := middleware.FirebaseUID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	// Resolve the Firestore user ID from the Firebase UID.
	userIter := h.Firestore.Collection("users").Where("firebase_uid", "==", firebaseUID).Limit(1).Documents(r.Context())
	userSnapshot, err := userIter.Next()
	userIter.Stop()
	if err == iterator.Done {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load user")
		return
	}
	var user models.User
	if err := userSnapshot.DataTo(&user); err != nil {
		writeError(w, http.StatusInternalServerError, "could not load user")
		return
	}
	user.ID = userSnapshot.Ref.ID

	// Fetch all bookings belonging to this user.
	var bookings []models.Booking
	iter := h.Firestore.Collection("bookings").Where("userID", "==", user.ID).Documents(r.Context())
	defer iter.Stop()

	for {
		snapshot, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not load bookings")
			return
		}

		var booking models.Booking
		if err := snapshot.DataTo(&booking); err != nil {
			writeError(w, http.StatusInternalServerError, "could not load bookings")
			return
		}
		booking.ID = snapshot.Ref.ID

		// Join the related slot document.
		slotSnapshot, err := h.Firestore.Collection("slots").Doc(booking.SlotID).Get(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not load bookings")
			return
		}
		var slot models.Slot
		if err := slotSnapshot.DataTo(&slot); err != nil {
			writeError(w, http.StatusInternalServerError, "could not load bookings")
			return
		}
		slot.ID = slotSnapshot.Ref.ID
		booking.Slot = &slot
		bookings = append(bookings, booking)
	}

	// Sort newest bookings first.
	sort.Slice(bookings, func(i, j int) bool {
		return bookings[i].CreatedAt.After(bookings[j].CreatedAt)
	})

	writeJSON(w, http.StatusOK, map[string]any{"bookings": bookings})
}

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

// parseIDField extracts a string ID from a json.RawMessage that may be
// encoded as either a JSON string ("5") or number (5).
func parseIDField(raw json.RawMessage, fieldName string) (string, error) {
	if len(raw) == 0 {
		return "", fmt.Errorf("%s is required", fieldName)
	}

	// Try numeric first.
	var numeric uint64
	if err := json.Unmarshal(raw, &numeric); err == nil {
		if numeric == 0 {
			return "", fmt.Errorf("%s must be greater than zero", fieldName)
		}
		return fmt.Sprintf("%d", numeric), nil
	}

	// Fall back to string.
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return "", fmt.Errorf("%s must be a number or string", fieldName)
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("%s is required", fieldName)
	}

	return text, nil
}
