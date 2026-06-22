// Package api is the single Vercel serverless function entry point.
// All /api/* requests are rewritten to this handler via vercel.json.
//
// Vercel requires:
//   - The file lives in the api/ directory
//   - The package is named "api"
//   - An exported Handler function with http.HandlerFunc signature
package api

import (
	"context"
	"log"
	"net/http"
	"sync"

	"bookly-backend/pkg/config"
	"bookly-backend/pkg/db"
	"bookly-backend/pkg/handlers"
	"bookly-backend/pkg/middleware"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/v4/auth"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// --------------------------------------------------------------------------
// Singleton state — initialised once per Vercel cold start.
// --------------------------------------------------------------------------

var (
	once   sync.Once
	router http.Handler
)

// Handler is the Vercel serverless function entry point.
// Every /api/* request hits this function after the vercel.json rewrite.
func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		router = mustInit()
	})
	router.ServeHTTP(w, r)
}

// --------------------------------------------------------------------------
// Router setup
// --------------------------------------------------------------------------

func mustInit() http.Handler {
	// ── Firebase initialisation ─────────────────────────────────────────
	authClient, firestoreClient, err := config.Init()
	if err != nil {
		log.Fatalf("firebase init: %v", err)
	}

	// Seed default slots on first deployment (idempotent).
	if err := db.SeedSlots(context.Background(), firestoreClient); err != nil {
		log.Printf("seed slots: %v", err)
	}

	return newRouter(authClient, firestoreClient)
}

func newRouter(authClient *auth.Client, firestoreClient *firestore.Client) http.Handler {
	r := chi.NewRouter()

	// ── Global middleware ───────────────────────────────────────────────
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"https://bookly-a33k.vercel.app",
			"https://booking-platform-943f9.web.app",
			"https://booking-platform-943f9.firebaseapp.com",
			"http://localhost:3000",
			"http://localhost:5500",
			"http://127.0.0.1:5500",
			"http://localhost:5501",
			"http://127.0.0.1:5501",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// ── 404 / 405 handlers ──────────────────────────────────────────────
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte(`{"error":"method not allowed"}`))
	})

	// ── Handler instances ───────────────────────────────────────────────
	authHandler := &handlers.AuthHandler{}
	slotsHandler := &handlers.SlotsHandler{Firestore: firestoreClient}
	usersHandler := &handlers.UsersHandler{Firestore: firestoreClient}
	bookingsHandler := &handlers.BookingsHandler{Firestore: firestoreClient}
	dataHandler := &handlers.DataHandler{Firestore: firestoreClient}
	authMiddleware := middleware.FirebaseAuth(authClient)

	// ── Public routes ───────────────────────────────────────────────────
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	r.Get("/api/slots", slotsHandler.List)

	// ── Protected routes ────────────────────────────────────────────────
	r.Group(func(protected chi.Router) {
		protected.Use(authMiddleware)

		// Auth verification
		protected.Post("/api/auth/verify", authHandler.Verify)

		// User management
		protected.Post("/api/users/sync", usersHandler.Sync)
		protected.Get("/api/user/{id}", usersHandler.GetByID)

		// Bookings
		protected.Post("/api/book", bookingsHandler.Create)
		protected.Get("/api/bookings", bookingsHandler.List)

		// Generic Firestore CRUD
		protected.Post("/api/data", dataHandler.Create)
		protected.Get("/api/data/{id}", dataHandler.Read)
		protected.Put("/api/data/{id}", dataHandler.Update)
		protected.Delete("/api/data/{id}", dataHandler.Delete)
	})

	return r
}
