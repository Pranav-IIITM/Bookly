const API_BASE = "http://localhost:8080";

import { getFreshIdToken, waitForAuthUser } from "./firebase-config.js";

const form = document.querySelector("#booking-form");
const slotSelect = document.querySelector("#slot");
const submitButton = document.querySelector("#booking-submit");
const statusMessage = document.querySelector("#booking-status");
const selectedSlotId = new URLSearchParams(window.location.search).get("slotId");

function setStatus(message, type = "") {
  statusMessage.textContent = message;
  statusMessage.className = `status-message ${type}`.trim();
}

function normalizeSlots(payload) {
  if (Array.isArray(payload)) {
    return payload;
  }

  if (Array.isArray(payload.slots)) {
    return payload.slots;
  }

  return [];
}

function slotId(slot) {
  return slot.id || slot.slotId || slot._id || slot.time || slot.label;
}

function slotLabel(slot) {
  const date = slot.date || slot.day || "";
  const time = slot.time || slot.startTime || slot.range || slot.label || slot.title || "";
  return [date, time].filter(Boolean).join(" - ") || `Slot ${slotId(slot) || ""}`.trim();
}

function renderSlotOptions(slots) {
  slotSelect.innerHTML = '<option value="">Select a slot</option>';

  slots.forEach((slot) => {
    const option = document.createElement("option");
    option.value = slotId(slot) || "";
    option.textContent = slotLabel(slot);
    option.selected = selectedSlotId && option.value === selectedSlotId;
    slotSelect.appendChild(option);
  });

  if (!slots.length) {
    slotSelect.innerHTML = '<option value="">No slots available</option>';
  }
}

async function loadSlots() {
  slotSelect.disabled = true;
  setStatus("Loading slots...");

  try {
    const response = await fetch(`${API_BASE}/api/slots`);

    if (!response.ok) {
      throw new Error(`Could not load slots. Server returned ${response.status}.`);
    }

    const data = await response.json();
    const slots = normalizeSlots(data);
    renderSlotOptions(slots);
    setStatus("Slots loaded.", "success");
  } catch (error) {
    renderSlotOptions([]);
    setStatus(error.message, "error");
  } finally {
    slotSelect.disabled = false;
  }
}

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  submitButton.disabled = true;
  setStatus("Submitting booking...");

  try {
    const token = await getFreshIdToken();
    const payload = {
      slotId: slotSelect.value
    };

    const response = await fetch(`${API_BASE}/api/book`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json"
      },
      body: JSON.stringify(payload)
    });

    if (!response.ok) {
      let message = `Booking failed. Server returned ${response.status}.`;

      try {
        const data = await response.json();
        message = data.error || message;
      } catch {
        // Keep the status-based message when the response body is not JSON.
      }

      throw new Error(message);
    }

    form.reset();
    setStatus("Booking confirmed.", "success");
  } catch (error) {
    setStatus(error.message, "error");
  } finally {
    submitButton.disabled = false;
  }
});

loadSlots();
