/**
 * confetti.js — Party popper + canvas confetti burst on booking confirmation.
 * Pure vanilla JS, no external libraries.
 * Watches #booking-status for the "Booking confirmed." message and triggers the show.
 */

(function () {
  "use strict";

  /* ── Config ─────────────────────────────────────────── */
  const PARTICLE_COUNT = 300;
  const GRAVITY        = 0.35;
  const FRICTION       = 0.985;
  const DURATION_MS    = 3500;  // how long confetti rains
  const SIDE_BURST     = 110;   // particles from each side cannon
  const COLORS = [
    "#0f766e", "#d97706", "#16764f", "#f59e0b",
    "#10b981", "#fbbf24", "#34d399", "#6ee7b7",
    "#ff6b6b", "#ff9f43", "#54a0ff", "#5f27cd",
    "#ee5a24", "#009432", "#0652dd", "#833471",
  ];

  /* ── DOM references ─────────────────────────────────── */
  const canvas     = document.getElementById("confetti-canvas");
  const popperL    = document.getElementById("popper-left");
  const popperR    = document.getElementById("popper-right");
  const statusEl   = document.getElementById("booking-status");

  if (!canvas || !statusEl) return;

  const ctx = canvas.getContext("2d");
  let particles = [];
  let animId    = null;
  let running   = false;

  /* ── Resize canvas to fill viewport ────────────────── */
  function resizeCanvas() {
    canvas.width  = window.innerWidth;
    canvas.height = window.innerHeight;
  }
  window.addEventListener("resize", resizeCanvas);
  resizeCanvas();

  /* ── Particle factory ───────────────────────────────── */
  function randomBetween(a, b) {
    return a + Math.random() * (b - a);
  }

  function randomColor() {
    return COLORS[Math.floor(Math.random() * COLORS.length)];
  }

  /**
   * Create a single particle.
   * @param {number} x  - origin x
   * @param {number} y  - origin y
   * @param {string} direction - "left" | "right" | "top"
   */
  function createParticle(x, y, direction) {
    const angle = direction === "left"
      ? randomBetween(-Math.PI * 0.15, Math.PI * 0.45)   // fan right-upward
      : direction === "right"
        ? randomBetween(Math.PI * 0.55, Math.PI * 1.15)  // fan left-upward
        : randomBetween(0, Math.PI * 2);

    const speed = randomBetween(7, 18);

    return {
      x,
      y,
      vx:      Math.cos(angle) * speed,
      vy:      -Math.abs(Math.sin(angle)) * speed - randomBetween(3, 8),
      color:   randomColor(),
      width:   randomBetween(7, 14),
      height:  randomBetween(4, 8),
      rotation: randomBetween(0, Math.PI * 2),
      rotationSpeed: randomBetween(-0.2, 0.2),
      opacity: 1,
      fade:    randomBetween(0.008, 0.018),
      shape:   Math.random() > 0.5 ? "rect" : "circle",
    };
  }

  function spawnBurst() {
    const W = canvas.width;
    const H = canvas.height;
    const midY = H * 0.45;

    // Left cannon burst
    for (let i = 0; i < SIDE_BURST; i++) {
      particles.push(createParticle(randomBetween(0, 60), randomBetween(midY - 40, midY + 40), "left"));
    }

    // Right cannon burst
    for (let i = 0; i < SIDE_BURST; i++) {
      particles.push(createParticle(randomBetween(W - 60, W), randomBetween(midY - 40, midY + 40), "right"));
    }

    // A few particles from the top center (extra delight)
    for (let i = 0; i < PARTICLE_COUNT - SIDE_BURST * 2; i++) {
      particles.push(createParticle(randomBetween(W * 0.2, W * 0.8), randomBetween(-20, 10), "top"));
    }
  }

  /* ── Draw loop ──────────────────────────────────────── */
  function drawParticle(p) {
    ctx.save();
    ctx.globalAlpha = Math.max(0, p.opacity);
    ctx.fillStyle   = p.color;
    ctx.translate(p.x, p.y);
    ctx.rotate(p.rotation);

    if (p.shape === "circle") {
      ctx.beginPath();
      ctx.arc(0, 0, p.width / 2, 0, Math.PI * 2);
      ctx.fill();
    } else {
      ctx.fillRect(-p.width / 2, -p.height / 2, p.width, p.height);
    }

    ctx.restore();
  }

  function tick() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);

    particles = particles.filter((p) => p.opacity > 0);

    particles.forEach((p) => {
      p.vy       += GRAVITY;
      p.vx       *= FRICTION;
      p.vy       *= FRICTION;
      p.x        += p.vx;
      p.y        += p.vy;
      p.rotation += p.rotationSpeed;
      p.opacity  -= p.fade;
      drawParticle(p);
    });

    if (particles.length > 0) {
      animId = requestAnimationFrame(tick);
    } else {
      stopConfetti();
    }
  }

  /* ── Popper helpers ─────────────────────────────────── */
  function showPoppers() {
    // Reset
    popperL.classList.remove("pop-out", "bounce");
    popperR.classList.remove("pop-out", "bounce");

    // Slide in
    requestAnimationFrame(() => {
      popperL.classList.add("pop-in");
      popperR.classList.add("pop-in");

      // Bounce after sliding in
      setTimeout(() => {
        popperL.classList.add("bounce");
        popperR.classList.add("bounce");
      }, 560);
    });
  }

  function hidePoppers(delay = 0) {
    setTimeout(() => {
      popperL.classList.remove("pop-in", "bounce");
      popperR.classList.remove("pop-in", "bounce");
      popperL.classList.add("pop-out");
      popperR.classList.add("pop-out");
    }, delay);
  }

  /* ── Main trigger ───────────────────────────────────── */
  function launchCelebration() {
    if (running) return;
    running = true;

    // Show canvas
    canvas.classList.add("active");

    // Spawn confetti
    spawnBurst();
    if (animId) cancelAnimationFrame(animId);
    animId = requestAnimationFrame(tick);

    // Party poppers
    showPoppers();

    // Auto-hide poppers after a bit
    hidePoppers(DURATION_MS - 800);

    // Stop after DURATION_MS even if particles remain
    setTimeout(stopConfetti, DURATION_MS);
  }

  function stopConfetti() {
    if (animId) {
      cancelAnimationFrame(animId);
      animId = null;
    }
    canvas.classList.remove("active");
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    particles = [];
    running   = false;
  }

  /* ── Watch #booking-status via MutationObserver ─────── */
  const SUCCESS_TEXTS = ["booking confirmed", "confirmed", "booked successfully", "success"];

  const observer = new MutationObserver(() => {
    const text = (statusEl.textContent || "").toLowerCase().trim();
    const isSuccess = statusEl.classList.contains("success") &&
      SUCCESS_TEXTS.some((t) => text.includes(t));

    if (isSuccess) {
      launchCelebration();
    }
  });

  observer.observe(statusEl, { childList: true, characterData: true, subtree: true, attributes: true });
})();
