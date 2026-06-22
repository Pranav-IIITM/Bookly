const API_BASE = "https://bookly-gules.vercel.app";

import { initializeApp } from "https://www.gstatic.com/firebasejs/9.23.0/firebase-app.js";
import {
  GoogleAuthProvider,
  createUserWithEmailAndPassword,
  getAuth,
  onAuthStateChanged,
  signInWithEmailAndPassword,
  signInWithPopup,
  signOut
} from "https://www.gstatic.com/firebasejs/9.23.0/firebase-auth.js";

const firebaseConfig = {
  apiKey: "AIzaSyCrYg2tHmJR8eperOLgndykPJF-RXoZwto",
  authDomain: "booking-platform-943f9.firebaseapp.com",
  projectId: "booking-platform-943f9",
  storageBucket: "booking-platform-943f9.firebasestorage.app",
  messagingSenderId: "646294293293",
  appId: "1:646294293293:web:fc37b5af8d8a3521ada1a4",
  measurementId: "G-TZG35ZWDMB"
};

const TOKEN_KEY = "firebaseIdToken";

export { API_BASE };

export const app = initializeApp(firebaseConfig);
export const auth = getAuth(app);
export const googleProvider = new GoogleAuthProvider();

export function saveToken(token) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}

export function getStoredToken() {
  return localStorage.getItem(TOKEN_KEY);
}

export async function getFreshIdToken() {
  if (!auth.currentUser) {
    throw new Error("Please sign in before continuing.");
  }

  const token = await auth.currentUser.getIdToken();
  saveToken(token);
  return token;
}

export function waitForAuthUser() {
  return new Promise((resolve) => {
    const unsubscribe = onAuthStateChanged(auth, async (user) => {
      unsubscribe();

      if (user) {
        saveToken(await user.getIdToken());
      }

      resolve(user);
    });
  });
}

export async function loginWithEmail(email, password) {
	const credential = await signInWithEmailAndPassword(auth, email, password);
	saveToken(await credential.user.getIdToken());
	return credential.user;
}

export async function signupWithEmail(email, password) {
  const credential = await createUserWithEmailAndPassword(auth, email, password);
  saveToken(await credential.user.getIdToken());
  return credential.user;
}

export async function loginWithGoogle() {
  const credential = await signInWithPopup(auth, googleProvider);
  saveToken(await credential.user.getIdToken());
  return credential.user;
}

export async function logoutUser() {
	await signOut(auth);
	clearToken();
}

export async function syncBackendUser(user = auth.currentUser) {
	if (!user) {
		throw new Error("Please sign in before continuing.");
	}

	const token = await user.getIdToken();
	saveToken(token);

	const response = await fetch(`${API_BASE}/api/users/sync`, {
		method: "POST",
		headers: {
			Authorization: `Bearer ${token}`,
			"Content-Type": "application/json"
		},
		body: JSON.stringify({
			name: user.displayName || "",
			email: user.email || ""
		})
	});

	if (!response.ok) {
		let message = `Could not sync user. Server returned ${response.status}.`;

		try {
			const data = await response.json();
			message = data.error || message;
		} catch {
			// Keep the status-based message when the response body is not JSON.
		}

		throw new Error(message);
	}

	return response.json();
}

onAuthStateChanged(auth, async (user) => {
  if (user) {
    saveToken(await user.getIdToken());
    return;
  }

  clearToken();
});
