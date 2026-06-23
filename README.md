<div align="center">

<img src="frontend/favicon.png" alt="Bookly Logo" width="80" height="80" />

# Bookly

**A full-stack appointment booking platform — pick a slot, sign in once, and your reservation is confirmed in real time.**

[![Go Version](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![Firebase](https://img.shields.io/badge/Firebase-Firestore%20%26%20Auth-FFCA28?style=flat-square&logo=firebase)](https://firebase.google.com/)
[![Deployed on Vercel](https://img.shields.io/badge/Backend-Vercel%20Serverless-000000?style=flat-square&logo=vercel)](https://vercel.com/)
[![Frontend on Vercel](https://img.shields.io/badge/Frontend-Vercel-000000?style=flat-square&logo=vercel)](https://vercel.com/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)

[Live Demo](https://bookly-5l61.vercel.app) · [API Health](https://bookly-5l61.vercel.app/api/health) · [Report a Bug](https://github.com/Pranav-IIITM/Bookly/issues)

</div>

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Firebase Setup](#firebase-setup)
  - [Backend — Local Development](#backend--local-development)
  - [Frontend — Local Development](#frontend--local-development)
- [Environment Variables](#environment-variables)
- [API Reference](#api-reference)
- [Deployment](#deployment)
  - [Backend — Vercel](#backend--vercel)
  - [Frontend — Firebase Hosting](#frontend--firebase-hosting)
- [Security](#security)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

Bookly is a modern appointment booking system built with a **Go backend** and a **vanilla HTML/CSS/JS frontend**. It uses Firebase Authentication for identity management and Cloud Firestore as its database.

Every booking goes through an **atomic Firestore transaction** — slot capacity is checked and the booking is created in a single operation, preventing double-bookings even under concurrent load. Both the frontend and backend are deployed on **Vercel** — the frontend as a static site and the backend as a serverless Go function.

---

## Features

| Feature | Description |
|---|---|
| 🔐 **Firebase Auth** | Sign in with Google OAuth or email/password. Firebase ID tokens are verified on every protected API call. |
| 📅 **Real-time Slot Availability** | Available slots are fetched live from the Go backend on every page load — no stale cache. |
| ⚛️ **Atomic Bookings** | Slot capacity checks and booking creation happen inside a single Firestore transaction, eliminating race conditions. |
| 🧾 **Personal Dashboard** | Authenticated users can view all their past and upcoming bookings, each enriched with the related slot details. |
| 🔄 **Dynamic Auth UI** | Navigation and hero buttons automatically reflect the user's sign-in state — "Sign In" becomes "Logout" after authentication. |
| 🎉 **Booking Confirmation** | A confetti animation fires on successful booking for a satisfying user experience. |
| ☁️ **Serverless Backend** | The Go API runs as a single Vercel serverless function with zero cold-start infrastructure to manage. |
| 🌱 **Auto-seeded Slots** | Default time slots are seeded automatically on first deployment via an idempotent seed function. |

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Browser / Client                      │
│  HTML + CSS + Vanilla JS  (Vercel Static)                   │
│                                                             │
│  ┌──────────────┐  Firebase SDK   ┌──────────────────────┐ │
│  │  Auth Pages  │ ◄────────────── │  Firebase Auth       │ │
│  └──────────────┘                 │  (ID Token issuance) │ │
│  ┌──────────────┐  REST + Bearer  └──────────────────────┘ │
│  │  Slots / Book│ ─────────────────────────────────────────┼─┐
│  │  Dashboard   │                                          │ │
│  └──────────────┘                                          │ │
└─────────────────────────────────────────────────────────────┘ │
                                                               │ HTTPS
                         ┌─────────────────────────────────────┘
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              Vercel Serverless Function (Go 1.22)            │
│  api/index.go  ──►  chi Router                              │
│                                                             │
│  Middleware: Firebase Token Verification, CORS, Logger      │
│                                                             │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────────┐  │
│  │ /api/slots  │  │  /api/book   │  │  /api/bookings    │  │
│  │ (public)    │  │  (protected) │  │  (protected)      │  │
│  └─────────────┘  └──────────────┘  └───────────────────┘  │
│                                                             │
│         Cloud Firestore (Firebase Admin SDK)                │
│   ┌──────────────┐  ┌───────────────┐  ┌────────────────┐  │
│   │   /slots     │  │   /bookings   │  │   /users       │  │
│   └──────────────┘  └───────────────┘  └────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

---

## Project Structure

```
Bookly/
│
├── api/
│   └── index.go              # Vercel serverless entry point; router setup & middleware
│
├── cmd/                      # Optional CLI tooling
│
├── pkg/
│   ├── config/               # Firebase Admin SDK initialisation (env vars / credentials)
│   ├── db/                   # Firestore seed functions
│   ├── handlers/
│   │   ├── auth.go           # POST /api/auth/verify
│   │   ├── bookings.go       # POST /api/book, GET /api/bookings
│   │   ├── data.go           # Generic CRUD: /api/data
│   │   ├── slots.go          # GET /api/slots
│   │   ├── users.go          # POST /api/users/sync, GET /api/user/:id
│   │   └── response.go       # Shared JSON/error helpers
│   ├── middleware/
│   │   └── auth.go           # Firebase ID-token verification middleware
│   └── models/               # Booking, Slot, User struct definitions
│
├── frontend/
│   ├── index.html            # Landing page
│   ├── auth.html             # Sign-in / sign-up page
│   ├── slots.html            # Available slots listing
│   ├── booking.html          # Booking confirmation page
│   ├── dashboard.html        # User's bookings dashboard
│   ├── favicon.png           # App icon
│   ├── css/
│   │   └── style.css         # Global design system & component styles
│   └── js/
│       ├── firebase-config.js # Firebase SDK initialisation (client-side)
│       ├── nav-auth.js        # Dynamic auth button state management
│       └── ...               # Page-specific scripts
│
├── main.go                   # Local development server entry point
├── go.mod / go.sum           # Go module definition & checksums
├── vercel.json               # Vercel routing & CORS header config
├── .env.example              # Template for required environment variables
└── .gitignore
```

---

## Tech Stack

### Backend
| | Technology | Purpose |
|---|---|---|
| 🐹 | **Go 1.22** | Core API language |
| 🔀 | **go-chi/chi v5** | HTTP router |
| 🔥 | **Firebase Admin SDK (Go)** | Firestore client & Auth token verification |
| ☁️ | **Cloud Firestore** | NoSQL database |
| ⚙️ | **joho/godotenv** | `.env` loading for local development |
| 🚀 | **Vercel** | Serverless deployment |

### Frontend
| | Technology | Purpose |
|---|---|---|
| 🌐 | **HTML5 / CSS3** | Markup and design system |
| ✨ | **Vanilla JavaScript (ESM)** | Page logic and API calls |
| 🔥 | **Firebase JS SDK v10** | Client-side auth (Google, email/password) |
| 🎉 | **canvas-confetti** | Booking confirmation animation |
| 🖋️ | **Inter (Google Fonts)** | Typography |
| 🏠 | **Vercel** | Static frontend hosting |

---

## Getting Started

### Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- A [Firebase project](https://console.firebase.google.com/) with **Firestore** and **Authentication** enabled
- A **Firebase service account** JSON key (download from *Project Settings → Service Accounts*)
- [Node.js](https://nodejs.org/) and [Firebase CLI](https://firebase.google.com/docs/cli) (for frontend deployment only)
- A local HTTP server such as [Live Server](https://marketplace.visualstudio.com/items?itemName=ritwickdey.LiveServer) for frontend development

---

### Firebase Setup

1. Go to [Firebase Console](https://console.firebase.google.com/) and create a project.
2. Enable **Authentication** → Sign-in methods → **Email/Password** and **Google**.
3. Enable **Firestore Database** in production mode.
4. Under *Project Settings → Service Accounts*, click **Generate new private key** and save the JSON file.
5. Register a **Web App** in Firebase Console and copy the config object — this goes into `frontend/js/firebase-config.js`.

---

### Backend — Local Development

**1. Clone the repository**
```bash
git clone https://github.com/Pranav-IIITM/Bookly.git
cd Bookly
```

**2. Configure environment variables**
```bash
cp .env.example .env
```

Open `.env` and fill in your values:
```env
PORT=8080
FIREBASE_PROJECT_ID=your-project-id

# Option A — Base64-encoded service account JSON (recommended)
FIREBASE_CREDENTIALS_JSON=<base64-encoded-service-account>

# Option B — Path to local JSON file (local dev only)
# FIREBASE_CREDENTIALS_PATH=pkg/firebase-credentials.json
```

To generate the Base64 value:
```powershell
# PowerShell
[Convert]::ToBase64String([IO.File]::ReadAllBytes("path\to\service-account.json"))
```
```bash
# Linux / macOS
base64 -w0 path/to/service-account.json
```

**3. Install Go dependencies**
```bash
go mod download
```

**4. Run the server**
```bash
go run main.go
# → Bookly backend listening on http://localhost:8080
```

The seed function will automatically populate Firestore with default slots if the `slots` collection is empty.

---

### Frontend — Local Development

1. Open `frontend/js/firebase-config.js` and ensure the Firebase config object matches your project.
2. Set the `API_BASE` variable to point to your local backend:
   ```js
   const API_BASE = "http://localhost:8080";
   ```
3. Serve the `frontend/` directory with any static server, e.g. VS Code Live Server on port `5500` or `5501`.
4. Navigate to `http://localhost:5500` (or whichever port your server uses).

---

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `PORT` | No | Port for the local server. Defaults to `8080`. |
| `FIREBASE_PROJECT_ID` | **Yes** | Your Firebase project ID (e.g. `my-project-abc12`). |
| `FIREBASE_CREDENTIALS_JSON` | **Yes*** | Base64-encoded Firebase service account JSON. Required in production (Vercel). |
| `FIREBASE_CREDENTIALS_PATH` | **Yes*** | Path to the service account JSON file. Alternative to `FIREBASE_CREDENTIALS_JSON` for local dev. |

> **\*** One of `FIREBASE_CREDENTIALS_JSON` or `FIREBASE_CREDENTIALS_PATH` must be provided.

---

## API Reference

All API responses are JSON. Protected endpoints require a Firebase ID token in the `Authorization: Bearer <token>` header.

### Public Endpoints

#### `GET /api/health`
Returns the server status and Firebase configuration info.

```json
{
  "status": "ok",
  "project_id": "your-project-id",
  "cred_length": 1234,
  "key_id": "abc123"
}
```

#### `GET /api/slots`
Returns all available (not fully booked) slots, sorted by date and time.

```json
{
  "slots": [
    {
      "id": "slot_001",
      "date": "2026-07-01",
      "time": "09:00",
      "capacity": 5,
      "bookedCount": 2
    }
  ]
}
```

---

### Protected Endpoints

> Require `Authorization: Bearer <firebase-id-token>` header.

#### `POST /api/auth/verify`
Verifies the provided Firebase token and returns the decoded claims.

```json
{
  "uid": "firebase_uid",
  "email": "user@example.com",
  "name": "Jane Doe"
}
```

#### `POST /api/users/sync`
Creates or updates the Firestore user document for the authenticated user. Call this after sign-in.

#### `GET /api/user/{id}`
Returns the user document for the given Firestore document ID.

#### `POST /api/book`
Creates a booking for the authenticated user. Atomically validates slot capacity.

**Request body:**
```json
{ "slotId": "slot_001" }
```

**Response (201):**
```json
{
  "booking": {
    "id": "booking_xyz",
    "userId": "user_abc",
    "slotId": "slot_001",
    "status": "confirmed",
    "createdAt": "2026-06-23T10:00:00Z",
    "slot": { ... }
  }
}
```

**Error responses:**
| Status | Reason |
|---|---|
| `400` | Slot is already full |
| `401` | Missing or invalid Firebase token |
| `404` | Slot or user not found |

#### `GET /api/bookings`
Returns all bookings for the authenticated user, newest first, each enriched with the related slot data.

```json
{
  "bookings": [ { ... } ]
}
```

#### Generic CRUD — `/api/data`
| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/data` | Create a Firestore document |
| `GET` | `/api/data/{id}` | Read a document by ID |
| `PUT` | `/api/data/{id}` | Update a document |
| `DELETE` | `/api/data/{id}` | Delete a document |

---

## Deployment

### Backend — Vercel

1. Push your code to GitHub.
2. Import the repository in the [Vercel dashboard](https://vercel.com/new).
3. Set the following **Environment Variables** in *Project Settings → Environment Variables*:
   - `FIREBASE_PROJECT_ID`
   - `FIREBASE_CREDENTIALS_JSON` (Base64-encoded service account)
4. Vercel will use `vercel.json` to route all `/api/*` requests to `api/index.go`.
5. Deploy — the Go function initialises Firebase once per cold start and caches the router.

### Frontend — Vercel

1. Push your code to GitHub (the `frontend/` directory is included in the repo root).
2. In the [Vercel dashboard](https://vercel.com/new), import the same repository.
3. Set the **Root Directory** to `frontend` in the project settings.
4. Update `API_BASE` in your JS files to point to your backend Vercel deployment URL.
5. Deploy — Vercel will serve the static files automatically on every push to `main`.

---

## Security

- **Credentials are never committed.** `.env`, all `*firebase-credentials*.json` files, and `base64_credentials.txt` are listed in `.gitignore`.
- **Firebase ID tokens are verified server-side** on every protected endpoint via the Firebase Admin SDK — the client cannot forge a valid token.
- **CORS is explicitly whitelisted** to only accept requests from the known frontend origins (Firebase Hosting and localhost).
- **Use `.env.example`** as the template. Never add real secrets to `.env.example`.

---

## Contributing

Contributions are welcome! Please follow these steps:

1. **Fork** the repository and create a feature branch:
   ```bash
   git checkout -b feat/your-feature-name
   ```
2. Make your changes and ensure the backend compiles:
   ```bash
   go build ./...
   ```
3. **Commit** using [Conventional Commits](https://www.conventionalcommits.org/):
   ```
   feat: add cancellation endpoint
   fix: handle empty slot collection gracefully
   ```
4. Open a **Pull Request** against `main` with a clear description of your changes.

Please open an issue first for significant features or breaking changes.

---

## License

This project is licensed under the [MIT License](LICENSE).

---

<div align="center">
  Built with ❤️ using Go, Firebase, and Vanilla JS.
</div>
