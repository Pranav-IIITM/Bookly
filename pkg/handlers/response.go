// Package handlers implements the HTTP handlers (controllers) for every
// Bookly API endpoint. Shared JSON helpers live in this file.
package handlers

import (
	"encoding/json"
	"net/http"
)

// writeJSON serialises payload as JSON and writes it with the given status
// code. Content-Type is always set to application/json.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// writeError is a convenience wrapper that writes a JSON error envelope:
//
//	{"error": "<message>"}
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
