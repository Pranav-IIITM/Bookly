package models

// Slot represents a bookable time slot. Capacity tracks the maximum number
// of bookings allowed, while BookedCount tracks how many have been made.
type Slot struct {
	ID          string `firestore:"-" json:"id"`
	Date        string `firestore:"date" json:"date"`
	Time        string `firestore:"time" json:"time"`
	Capacity    int    `firestore:"capacity" json:"capacity"`
	BookedCount int    `firestore:"bookedCount" json:"bookedCount"`
}
