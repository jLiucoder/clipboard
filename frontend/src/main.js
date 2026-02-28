import { Events } from "@wailsio/runtime";
import { App } from "../bindings/changeme";

const island = document.getElementById("island");
const islandBody = document.getElementById("island-body");
const islandCount = document.getElementById("island-count");

let isOpen = false;
let selectedIndex = -1;
let allItems = [];

// ── Render clipboard history ─────────────────────────────────────────────────
function renderHistory(items) {
  // Clear current content
  islandBody.innerHTML = "";
  allItems = items;

  if (items.length === 0) {
    const empty = document.createElement("div");
    empty.className = "empty-state";
    empty.textContent = "Copy text to get started";
    islandBody.appendChild(empty);
    islandCount.textContent = "0";
    selectedIndex = -1;
    return;
  }

  // Create list container
  const list = document.createElement("div");
  list.className = "clip-list";

  // Add each item
  items.forEach((item, index) => {
    const row = document.createElement("div");
    row.className = "clip-row" + (item.pinned ? " pinned" : "");
    row.dataset.index = index;

    const text = document.createElement("div");
    text.className = "clip-text";
    text.textContent = item.text;

    // Action buttons container
    const actions = document.createElement("div");
    actions.className = "clip-actions";

    // Pin button (☆/★)
    const pinBtn = document.createElement("button");
    pinBtn.className = "clip-btn pin-btn";
    pinBtn.textContent = item.pinned ? "★" : "☆";
    pinBtn.title = item.pinned ? "Unpin" : "Pin";
    pinBtn.addEventListener("click", (e) => {
      e.stopPropagation();
      togglePin(index);
    });

    // Delete button (×)
    const delBtn = document.createElement("button");
    delBtn.className = "clip-btn del-btn";
    delBtn.textContent = "×";
    delBtn.title = "Delete";
    delBtn.addEventListener("click", (e) => {
      e.stopPropagation();
      deleteItem(index);
    });

    actions.appendChild(pinBtn);
    actions.appendChild(delBtn);

    row.appendChild(text);
    row.appendChild(actions);
    list.appendChild(row);

    // Click on row to paste
    row.addEventListener("click", () => {
      selectAndPaste(index);
    });

    // Mouse hover updates selection
    row.addEventListener("mouseenter", () => {
      updateSelection(index);
    });
  });

  islandBody.appendChild(list);
  islandCount.textContent = String(items.length);

  // Select first item by default
  if (selectedIndex < 0 || selectedIndex >= items.length) {
    selectedIndex = 0;
  }
  updateSelection(selectedIndex);
}

// ── Update selection highlight ───────────────────────────────────────────────
function updateSelection(index) {
  if (index < 0 || index >= allItems.length) return;

  selectedIndex = index;

  // Update visual selection
  const rows = islandBody.querySelectorAll(".clip-row");
  rows.forEach((row, i) => {
    if (i === selectedIndex) {
      row.classList.add("selected");
      // Scroll into view if needed
      row.scrollIntoView({ block: "nearest", behavior: "smooth" });
    } else {
      row.classList.remove("selected");
    }
  });
}

// ── Toggle pin status ─────────────────────────────────────────────────────────
async function togglePin(index) {
  try {
    await App.TogglePin(index);
    await refreshHistory();
  } catch (err) {
    console.error("Failed to toggle pin:", err);
  }
}

// ── Delete item ───────────────────────────────────────────────────────────────
async function deleteItem(index) {
  try {
    await App.DeleteItem(index);
    await refreshHistory();
  } catch (err) {
    console.error("Failed to delete item:", err);
  }
}

// ── Select and paste item ────────────────────────────────────────────────────
async function selectAndPaste(index) {
  if (index < 0 || index >= allItems.length) return;
  
  try {
    isOpen = false;
    island.classList.remove("open");
    await App.SelectItem(index);
  } catch (err) {
    console.error("Failed to paste:", err);
  }
}

// ── Fetch and render history ─────────────────────────────────────────────────
async function refreshHistory() {
  try {
    const history = await App.GetHistory();
    renderHistory(history);
  } catch (err) {
    console.error("Failed to get history:", err);
  }
}

// ── Hotkey event from Go ──────────────────────────────────────────────────────
Events.On("hotkey", () => {
  island.classList.remove("open");
  void island.offsetHeight;
  island.classList.add("open");
  isOpen = true;
  selectedIndex = 0;
  window.focus();
  refreshHistory();
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

  if (e.key === "ArrowDown") {
    e.preventDefault();
    const nextIndex = selectedIndex + 1;
    if (nextIndex < allItems.length) {
      updateSelection(nextIndex);
    } else {
      // Wrap to top
      updateSelection(0);
    }
    return;
  }

  if (e.key === "ArrowUp") {
    e.preventDefault();
    const prevIndex = selectedIndex - 1;
    if (prevIndex >= 0) {
      updateSelection(prevIndex);
    } else {
      // Wrap to bottom
      updateSelection(allItems.length - 1);
    }
    return;
  }

  if (e.key === "Enter") {
    e.preventDefault();
    if (selectedIndex >= 0 && selectedIndex < allItems.length) {
      selectAndPaste(selectedIndex);
    }
    return;
  }
});
