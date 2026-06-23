const API_BASE = "https://bookly-gules.vercel.app";

const slotsList = document.querySelector("#slots-list");
const refreshButton = document.querySelector("#refresh-slots");
const statusMessage = document.querySelector("#slots-status");

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

function slotTitle(slot) {
  return slot.label || slot.title || slot.time || `Slot ${slotId(slot) || ""}`.trim();
}

function slotDate(slot) {
  return slot.date || slot.day || slot.startDate || "Date to be confirmed";
}

function slotTime(slot) {
  return slot.time || slot.startTime || slot.range || "Time to be confirmed";
}

function sortSlots(slots) {
  return [...slots].sort((a, b) => {
    const toDate = (slot) => {
      const datePart = slotDate(slot);
      const timePart = slotTime(slot);
      // Combine into a single string that Date.parse can handle
      const combined = `${datePart} ${timePart}`;
      const parsed = Date.parse(combined);
      // Fall back to comparing the raw strings if parsing fails
      return isNaN(parsed) ? combined : parsed;
    };

    const da = toDate(a);
    const db = toDate(b);

    if (typeof da === "number" && typeof db === "number") {
      return da - db;
    }
    // Lexicographic fallback for unparseable formats
    return String(da).localeCompare(String(db));
  });
}

function renderSlots(slots) {
  slotsList.innerHTML = "";

  if (!slots.length) {
    const emptyState = document.createElement("div");
    const heading = document.createElement("h3");
    const copy = document.createElement("p");

    emptyState.className = "empty-state";
    heading.textContent = "No slots available";
    copy.textContent = "Try refreshing or check again later.";
    emptyState.append(heading, copy);
    slotsList.appendChild(emptyState);
    return;
  }

  const fragment = document.createDocumentFragment();

  slots.forEach((slot) => {
    const id = slotId(slot);
    const article = document.createElement("article");
    const header = document.createElement("div");
    const eyebrow = document.createElement("p");
    const title = document.createElement("h3");
    const meta = document.createElement("div");
    const date = document.createElement("p");
    const time = document.createElement("p");
    const link = document.createElement("a");

    article.className = "slot-card";
    eyebrow.className = "eyebrow";
    eyebrow.textContent = "Available";
    title.textContent = slotTitle(slot);
    header.append(eyebrow, title);

    meta.className = "slot-meta";
    date.append("Date: ", slotDate(slot));
    time.append("Time: ", slotTime(slot));
    meta.append(date, time);

    link.className = "button";
    link.href = `booking.html?slotId=${encodeURIComponent(id || "")}`;
    link.textContent = "Book";

    article.append(header, meta, link);
    fragment.appendChild(article);
  });

  slotsList.appendChild(fragment);
}

async function fetchSlots() {
  refreshButton.disabled = true;
  slotsList.innerHTML = "";
  setStatus("Loading available slots...");

  try {
    const response = await fetch(`${API_BASE}/api/slots`);

    if (!response.ok) {
      throw new Error(`Could not load slots. Server returned ${response.status}.`);
    }

    const data = await response.json();
    const slots = sortSlots(normalizeSlots(data));
    renderSlots(slots);
    setStatus(`${slots.length} slot${slots.length === 1 ? "" : "s"} loaded.`, "success");
  } catch (error) {
    renderSlots([]);
    setStatus(error.message, "error");
  } finally {
    refreshButton.disabled = false;
  }
}

refreshButton.addEventListener("click", fetchSlots);
fetchSlots();
