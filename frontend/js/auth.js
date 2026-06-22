const API_BASE = "http://localhost:8080";

import {
	auth,
	loginWithEmail,
	loginWithGoogle,
	logoutUser,
	signupWithEmail,
	syncBackendUser
} from "./firebase-config.js";
import { onAuthStateChanged } from "https://www.gstatic.com/firebasejs/9.23.0/firebase-auth.js";

const form = document.querySelector("#auth-form");
const emailInput = document.querySelector("#email");
const passwordInput = document.querySelector("#password");
const submitButton = document.querySelector("#auth-submit");
const googleButton = document.querySelector("#google-login");
const logoutButton = document.querySelector("#logout-button");
const statusMessage = document.querySelector("#auth-status");
const modeButtons = document.querySelectorAll("[data-auth-mode]");

let authMode = "login";

function setStatus(message, type = "") {
  statusMessage.textContent = message;
  statusMessage.className = `status-message ${type}`.trim();
}

function setLoading(isLoading) {
  submitButton.disabled = isLoading;
  googleButton.disabled = isLoading;
  logoutButton.disabled = isLoading;
}

function updateMode(mode) {
  authMode = mode;
  submitButton.textContent = mode === "login" ? "Login" : "Sign Up";
  passwordInput.autocomplete = mode === "login" ? "current-password" : "new-password";

  modeButtons.forEach((button) => {
    button.classList.toggle("active", button.dataset.authMode === mode);
  });
}

modeButtons.forEach((button) => {
  button.addEventListener("click", () => updateMode(button.dataset.authMode));
});

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  setLoading(true);
  setStatus(authMode === "login" ? "Signing you in..." : "Creating your account...");

	try {
		const email = emailInput.value.trim();
		const password = passwordInput.value;
		let user;

		if (authMode === "login") {
			user = await loginWithEmail(email, password);
			setStatus("Signed in successfully.", "success");
		} else {
			user = await signupWithEmail(email, password);
			setStatus("Account created successfully.", "success");
		}

		await syncBackendUser(user);
		window.location.href = "dashboard.html";
	} catch (error) {
		setStatus(error.message, "error");
  } finally {
    setLoading(false);
  }
});

googleButton.addEventListener("click", async () => {
  setLoading(true);
	setStatus("Opening Google sign-in...");

	try {
		const user = await loginWithGoogle();
		await syncBackendUser(user);
		setStatus("Signed in successfully.", "success");
		window.location.href = "dashboard.html";
	} catch (error) {
    setStatus(error.message, "error");
  } finally {
    setLoading(false);
  }
});

logoutButton.addEventListener("click", async () => {
  setLoading(true);
  setStatus("Signing out...");

  try {
    await logoutUser();
    setStatus("Signed out.", "success");
  } catch (error) {
    setStatus(error.message, "error");
  } finally {
    setLoading(false);
  }
});

onAuthStateChanged(auth, (user) => {
  logoutButton.classList.toggle("hidden", !user);
});

void API_BASE;
