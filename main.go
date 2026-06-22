// Bookly Backend — Local Development Entry Point
//
// This file is used for running the backend locally with `go run main.go`.
// Vercel ignores this file; it uses api/index.go instead.
//
// Usage:
//   1. Copy .env.example to .env and fill in your Firebase credentials
//   2. go run main.go
//   3. The server starts on http://localhost:8080
package main

import (
	"context"
	"log"
	"net/http"

	"bookly-backend/pkg/config"
	"bookly-backend/pkg/db"

	// Re-use the Vercel handler so all routes are identical.
	api "bookly-backend/api"
)

func main() {
	// Init Firebase (also loads .env via godotenv).
	_, firestoreClient, err := config.Init()
	if err != nil {
		log.Fatalf("firebase init: %v", err)
	}
	defer firestoreClient.Close()

	// Seed default slots if the collection is empty.
	if err := db.SeedSlots(context.Background(), firestoreClient); err != nil {
		log.Printf("seed slots: %v", err)
	}

	port := config.GetEnv("PORT", "8080")
	addr := ":" + port

	log.Printf("Bookly backend listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, http.HandlerFunc(api.Handler)); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
