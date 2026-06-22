// Package config handles Firebase Admin SDK initialisation and environment
// variable loading. It uses sync.Once so the expensive Firebase bootstrap
// only runs once per cold start in Vercel's serverless runtime.
package config

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"sync"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// --------------------------------------------------------------------------
// Singleton state — initialised once per cold start.
// --------------------------------------------------------------------------

var (
	once            sync.Once
	initErr         error
	authClient      *auth.Client
	firestoreClient *firestore.Client
)

// Init bootstraps the Firebase Admin SDK. It is safe to call from multiple
// goroutines; the heavy work only runs once.
func Init() (*auth.Client, *firestore.Client, error) {
	once.Do(func() {
		// Attempt to load a .env file for local development. In production
		// (Vercel) the variables are set via the dashboard, so this is a
		// no-op if the file is missing.
		_ = godotenv.Load()

		authClient, firestoreClient, initErr = bootstrap()
	})

	return authClient, firestoreClient, initErr
}

// --------------------------------------------------------------------------
// Internal helpers
// --------------------------------------------------------------------------

// bootstrap creates the Firebase app, auth client, and Firestore client.
func bootstrap() (*auth.Client, *firestore.Client, error) {
	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		return nil, nil, errors.New("FIREBASE_PROJECT_ID is required")
	}

	// Resolve credentials: either base64-encoded JSON in an env var (Vercel)
	// or a file path on disk (local dev).
	opts, err := credentialOption()
	if err != nil {
		return nil, nil, fmt.Errorf("firebase credentials: %w", err)
	}

	cfg := &firebase.Config{ProjectID: projectID}
	app, err := firebase.NewApp(context.Background(), cfg, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("firebase app: %w", err)
	}

	ac, err := app.Auth(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("firebase auth: %w", err)
	}

	fc, err := app.Firestore(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("firestore: %w", err)
	}

	return ac, fc, nil
}

// credentialOption returns the google.golang.org/api/option for configuring
// Firebase credentials. It prefers FIREBASE_CREDENTIALS_JSON (base64) and
// falls back to FIREBASE_CREDENTIALS_PATH (file).
func credentialOption() ([]option.ClientOption, error) {
	// 1. Base64-encoded JSON (Vercel)
	if encoded := os.Getenv("FIREBASE_CREDENTIALS_JSON"); encoded != "" {
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("decoding FIREBASE_CREDENTIALS_JSON: %w", err)
		}
		return []option.ClientOption{option.WithCredentialsJSON(decoded)}, nil
	}

	// 2. Direct credentials file
	return []option.ClientOption{option.WithCredentialsFile("pkg/backendfirebase-credentials.json")}, nil
}

// GetEnv returns the value of an environment variable or a fallback default.
func GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
