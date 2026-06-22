// Package models defines the Firestore document structures used throughout the
// Bookly backend. Struct tags map fields to both Firestore documents and JSON
// API responses.
package models

import "time"

// User represents a registered user synced from Firebase Authentication.
// The ID field is the Firestore document ID and is not stored inside the
// document itself (firestore:"-").
type User struct {
	ID          string    `firestore:"-" json:"id"`
	FirebaseUID string    `firestore:"firebase_uid" json:"firebaseUid"`
	Name        string    `firestore:"name" json:"name"`
	Email       string    `firestore:"email" json:"email"`
	CreatedAt   time.Time `firestore:"createdAt" json:"createdAt"`
}
