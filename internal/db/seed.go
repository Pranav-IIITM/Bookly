// Package db provides database seeding utilities for initial setup.
package db

import (
	"context"

	"bookly-backend/internal/models"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// SeedSlots populates the "slots" collection with default time slots if it
// is currently empty. This runs once on first startup so the frontend has
// data to display immediately.
func SeedSlots(ctx context.Context, client *firestore.Client) error {
	slotsCol := client.Collection("slots")

	// Check whether the collection already has at least one document.
	iter := slotsCol.Limit(1).Documents(ctx)
	_, err := iter.Next()
	iter.Stop()

	if err == nil {
		// Collection is not empty — skip seeding.
		return nil
	}
	if err != iterator.Done {
		return err
	}

	// ── Seed data ───────────────────────────────────────────────────────
	slots := map[string]models.Slot{
		"1": {Date: "2026-07-01", Time: "10:00 AM", Capacity: 10, BookedCount: 0},
		"2": {Date: "2026-07-01", Time: "12:00 PM", Capacity: 10, BookedCount: 0},
		"3": {Date: "2026-07-02", Time: "10:00 AM", Capacity: 5, BookedCount: 0},
		"4": {Date: "2026-07-02", Time: "02:00 PM", Capacity: 8, BookedCount: 0},
		"5": {Date: "2026-07-03", Time: "11:00 AM", Capacity: 6, BookedCount: 0},
	}

	for id, slot := range slots {
		if _, err := slotsCol.Doc(id).Set(ctx, slot); err != nil {
			return err
		}
	}

	return nil
}
