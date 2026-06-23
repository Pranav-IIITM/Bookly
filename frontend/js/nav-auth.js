import { auth, logoutUser } from "./firebase-config.js";
import { onAuthStateChanged } from "https://www.gstatic.com/firebasejs/9.23.0/firebase-auth.js";

/**
 * Updates the nav auth button and any optional hero CTA based on auth state.
 *
 * Expected elements (by ID) – all optional, the script is safe if they're absent:
 *   #nav-auth-btn      – the header "Sign In / Logout" button/link
 *   #hero-auth-btn     – the hero "Sign In to Book / Go to Dashboard" CTA (index.html only)
 */
function applyAuthState(user) {
  const navBtn = document.getElementById("nav-auth-btn");
  const heroBtn = document.getElementById("hero-auth-btn");

  if (user) {
    // --- Signed in ---
    if (navBtn) {
      navBtn.textContent = "Logout";
      navBtn.removeAttribute("href");
      navBtn.style.cursor = "pointer";
      navBtn.onclick = async () => {
        await logoutUser();
        // Redirect to home after logout so UI resets cleanly
        window.location.href = "index.html";
      };
    }

    if (heroBtn) {
      heroBtn.textContent = "Go to Dashboard";
      heroBtn.setAttribute("href", "dashboard.html");
      heroBtn.onclick = null;
    }
  } else {
    // --- Signed out ---
    if (navBtn) {
      navBtn.textContent = "Sign In";
      navBtn.setAttribute("href", "auth.html");
      navBtn.style.cursor = "";
      navBtn.onclick = null;
    }

    if (heroBtn) {
      heroBtn.textContent = "Sign In to Book";
      heroBtn.setAttribute("href", "auth.html");
      heroBtn.onclick = null;
    }
  }
}

onAuthStateChanged(auth, applyAuthState);
