package models

import "time"

// Booking ties a User to a Slot. The Slot pointer is populated at read time
// by joining on SlotID; it is never persisted to Firestore (firestore:"-").
type Booking struct {
	ID        string    `firestore:"-" json:"id"`
	UserID    string    `firestore:"userID" json:"userId"`
	SlotID    string    `firestore:"slotID" json:"slotId"`
	Status    string    `firestore:"status" json:"status"`
	CreatedAt time.Time `firestore:"createdAt" json:"createdAt"`
	Slot      *Slot     `firestore:"-" json:"slot,omitempty"`
}
