const API_BASE = "https://bookly-gules.vercel.app";

import { getFreshIdToken, logoutUser, waitForAuthUser } from "./firebase-config.js";

const bookingsList = document.querySelector("#bookings-list");
const refreshButton = document.querySelector("#refresh-bookings");
const logoutButton = document.querySelector("#logout-button");
const profileLogoutButton = document.querySelector("#profile-logout");
const statusMessage = document.querySelector("#dashboard-status");

// ── Profile helpers ──────────────────────────────────────────

const PROFILE_KEY = "booklyUserProfile";

/** Persist user profile data to localStorage. */
function saveProfile(user) {
  const profile = {
    name: user.displayName || "",
    email: user.email || ""
  };
  localStorage.setItem(PROFILE_KEY, JSON.stringify(profile));
  return profile;
}

/** Retrieve stored profile from localStorage. */
function loadProfile() {
  try {
    const raw = localStorage.getItem(PROFILE_KEY);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

/** Clear stored profile from localStorage. */
function clearProfile() {
  localStorage.removeItem(PROFILE_KEY);
}

/**
 * Generate initials from a full name (up to 2 characters).
 * Falls back to the first character of the email if no name is set.
 */
function getInitials(name, email) {
  if (name && name.trim()) {
    const parts = name.trim().split(/\s+/);
    if (parts.length >= 2) {
      return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
    }
    return parts[0].slice(0, 2).toUpperCase();
  }
  if (email) {
    return email[0].toUpperCase();
  }
  return "?";
}

/**
 * Derive a consistent HSL background color from a string (name or email).
 * Uses a simple DJB2-style hash so the same user always gets the same color.
 */
function avatarColor(str) {
  let hash = 5381;
  for (let i = 0; i < str.length; i++) {
    hash = ((hash << 5) + hash) ^ str.charCodeAt(i);
  }
  // Keep hue in the 180-340 range (cool/teal to purple) for contrast with white text.
  const hue = Math.abs(hash) % 160 + 180;
  return `hsl(${hue}, 55%, 38%)`;
}

/** Render the profile card if a profile is available. */
function renderProfile(profile) {
  const section = document.querySelector("#user-profile");
  const avatarEl = document.querySelector("#profile-avatar");
  const nameEl = document.querySelector("#profile-name");
  const emailEl = document.querySelector("#profile-email");

  if (!profile || (!profile.name && !profile.email)) {
    return;
  }

  const initials = getInitials(profile.name, profile.email);
  const color = avatarColor(profile.name || profile.email);

  avatarEl.textContent = initials;
  avatarEl.style.background = color;
  nameEl.textContent = profile.name || "Signed-in user";
  emailEl.textContent = profile.email || "";

  section.hidden = false;
}

// ── Status helpers ───────────────────────────────────────────

function setStatus(message, type = "") {
  statusMessage.textContent = message;
  statusMessage.className = `status-message ${type}`.trim();
}

// ── Booking helpers ──────────────────────────────────────────

function normalizeBookings(payload) {
  if (Array.isArray(payload)) {
    return payload;
  }

  if (Array.isArray(payload.bookings)) {
    return payload.bookings;
  }

  return [];
}

function bookingTitle(booking, index) {
  const slot = booking.slot || {};
  return booking.title || booking.service || booking.slotLabel || slot.time || `Booking ${index + 1}`;
}

function renderBookings(bookings) {
  bookingsList.innerHTML = "";

  if (!bookings.length) {
    const emptyState = document.createElement("div");
    const heading = document.createElement("h3");
    const copy = document.createElement("p");

    emptyState.className = "empty-state";
    heading.textContent = "No bookings yet";
    copy.textContent = "Your confirmed reservations will appear here.";
    emptyState.append(heading, copy);
    bookingsList.appendChild(emptyState);
    return;
  }

  const fragment = document.createDocumentFragment();

  bookings.forEach((booking, index) => {
    const bookedSlot = booking.slot || {};
    const article = document.createElement("article");
    const header = document.createElement("div");
    const eyebrow = document.createElement("p");
    const title = document.createElement("h3");
    const meta = document.createElement("div");
    const name = document.createElement("p");
    const date = document.createElement("p");
    const slot = document.createElement("p");

    article.className = "booking-card";
    eyebrow.className = "eyebrow";
    eyebrow.textContent = "Confirmed";
    title.textContent = bookingTitle(booking, index);
    header.append(eyebrow, title);

    meta.className = "booking-meta";
    name.append("Status: ", booking.status || "confirmed");
    date.append("Date: ", booking.date || booking.day || bookedSlot.date || "Date to be confirmed");
    slot.append("Slot: ", bookedSlot.time || booking.slotId || booking.time || "Slot to be confirmed");
    meta.append(name, date, slot);

    article.append(header, meta);
    fragment.appendChild(article);
  });

  bookingsList.appendChild(fragment);
}

// ── Auth & data loading ──────────────────────────────────────

async function handleLogout() {
  clearProfile();
  await logoutUser();
  window.location.href = "auth.html";
}

async function fetchBookings() {
  refreshButton.disabled = true;
  bookingsList.innerHTML = "";
  setStatus("Loading your bookings...");

  try {
    const user = await waitForAuthUser();

    if (!user) {
      window.location.href = "auth.html";
      return;
    }

    // Persist and render profile on every load so it stays fresh.
    const profile = saveProfile(user);
    renderProfile(profile);

    const token = await getFreshIdToken();
    const response = await fetch(`${API_BASE}/api/bookings`, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });

    if (!response.ok) {
      throw new Error(`Could not load bookings. Server returned ${response.status}.`);
    }

    const data = await response.json();
    const bookings = normalizeBookings(data);
    renderBookings(bookings);
    setStatus(`${bookings.length} booking${bookings.length === 1 ? "" : "s"} loaded.`, "success");
  } catch (error) {
    renderBookings([]);
    setStatus(error.message, "error");
  } finally {
    refreshButton.disabled = false;
  }
}

// ── Event wiring ─────────────────────────────────────────────

refreshButton.addEventListener("click", fetchBookings);
logoutButton.addEventListener("click", handleLogout);
profileLogoutButton.addEventListener("click", handleLogout);

// Show cached profile immediately while Firebase initialises (avoids flash).
const cachedProfile = loadProfile();
if (cachedProfile) {
  renderProfile(cachedProfile);
}

fetchBookings();

