import { Events } from "@wailsio/runtime";
import { App } from "../bindings/changeme";

const island = document.getElementById("island");

let isOpen = false;

// ── Hotkey event from Go ──────────────────────────────────────────────────────
Events.On("hotkey", () => {
  island.classList.remove("open");
  void island.offsetHeight;
  island.classList.add("open");
  isOpen = true;
  window.focus();
});

// ── Dismiss helper ────────────────────────────────────────────────────────────
function dismiss() {
  if (!isOpen) return;
  isOpen = false;
  island.classList.remove("open");
  island.addEventListener(
    "transitionend",
    () => App.HideWindow().catch(() => {}),
    { once: true }
  );
}

// ── Keyboard ──────────────────────────────────────────────────────────────────
document.addEventListener("keydown", (e) => {
  if (!isOpen) return;

  if (e.key === "Escape") {
    e.preventDefault();
    dismiss();
    return;
  }
});
