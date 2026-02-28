package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"golang.design/x/clipboard"
)

// ClipItem represents a single clipboard item.
type ClipItem struct {
	Text   string `json:"text"`
	Pinned bool   `json:"pinned"`
}

// initClipboard initializes the clipboard package.
func (a *App) initClipboard() error {
	return clipboard.Init()
}

// watchClipboard polls the clipboard for changes and captures content.
func (a *App) watchClipboard() {
	log.Println("[clipboard] Starting clipboard watcher...")

	// Initial change count
	lastCount := getPasteboardChangeCount()
	lastText := ""

	for {
		time.Sleep(200 * time.Millisecond)

		currentCount := getPasteboardChangeCount()
		if currentCount == lastCount {
			continue
		}

		// Change detected, read clipboard
		data := clipboard.Read(clipboard.FmtText)
		if len(data) == 0 {
			lastCount = currentCount
			continue
		}

		text := string(data)

		// Skip if same as last captured text (prevents duplicates from rapid polling)
		if text == lastText {
			lastCount = currentCount
			continue
		}

		lastText = text
		lastCount = currentCount

		a.addItem(text)
		log.Printf("[clipboard] Captured %d chars", len(text))
	}
}

// addItem adds a new item to the clipboard history.
// It prepends to the front, dedups (moves to top), caps at 30, and skips empty/whitespace.
// It also skips items that match lastWritten (to avoid re-capturing pasted content).
// If same text exists and is pinned, the new item is skipped (don't re-add).
func (a *App) addItem(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Skip if this is the text we just wrote (from SelectItem)
	if text == a.lastWritten {
		a.lastWritten = "" // Clear after one skip
		return
	}

	// Check for duplicates
	for i, item := range a.history {
		if item.Text == text {
			// If existing item is pinned, skip the new addition entirely
			if item.Pinned {
				return
			}
			// Remove existing non-pinned item (will be re-added at front)
			a.history = append(a.history[:i], a.history[i+1:]...)
			break
		}
	}

	// Add new item at the front
	newItem := ClipItem{Text: text, Pinned: false}
	a.history = append([]ClipItem{newItem}, a.history...)

	// Cap at 30 items, but preserve pinned items
	if len(a.history) > 30 {
		a.history = a.trimToCap()
	}
}

// trimToCap reduces history to 30 items while preserving pinned items.
// It removes oldest non-pinned items first.
func (a *App) trimToCap() []ClipItem {
	if len(a.history) <= 30 {
		return a.history
	}

	var result []ClipItem
	var nonPinned []ClipItem

	// Separate pinned and non-pinned
	for _, item := range a.history {
		if item.Pinned {
			result = append(result, item)
		} else {
			nonPinned = append(nonPinned, item)
		}
	}

	// Add non-pinned items from the front until we hit the cap
	spaceForNonPinned := 30 - len(result)
	if spaceForNonPinned > 0 && len(nonPinned) > 0 {
		if len(nonPinned) > spaceForNonPinned {
			nonPinned = nonPinned[:spaceForNonPinned]
		}
		result = append(nonPinned, result...)
	}

	return result
}

// GetHistory returns a copy of the clipboard history.
// This is exported for Wails binding.
func (a *App) GetHistory() []ClipItem {
	a.mu.Lock()
	defer a.mu.Unlock()

	result := make([]ClipItem, len(a.history))
	copy(result, a.history)
	return result
}

// SelectItem selects an item from history, copies it to clipboard, hides the window,
// restores focus to the previous app, and simulates paste.
func (a *App) SelectItem(index int) {
	a.mu.Lock()
	if index < 0 || index >= len(a.history) {
		a.mu.Unlock()
		log.Printf("[clipboard] SelectItem: invalid index %d", index)
		return
	}
	text := a.history[index].Text
	a.lastWritten = text
	a.mu.Unlock()

	// Write to system clipboard
	clipboard.Write(clipboard.FmtText, []byte(text))

	// Update change count to prevent re-capturing
	a.lastChangeCount = getPasteboardChangeCount()

	// Hide window
	a.window.Hide()

	// Restore focus and paste in background
	go func() {
		time.Sleep(50 * time.Millisecond)
		a.restorePreviousApp()
		time.Sleep(100 * time.Millisecond)
		simulatePaste()
	}()
}

// simulatePaste simulates Cmd+V keystroke using AppleScript.
func simulatePaste() {
	script := `tell application "System Events" to keystroke "v" using command down`
	if err := exec.Command("osascript", "-e", script).Run(); err != nil {
		log.Printf("[clipboard] simulatePaste failed: %v", err)
	}
}

// getHistoryFilePath returns the path to the history file.
func getHistoryFilePath() string {
	return filepath.Join(xdg.DataHome, "clipboard-island", "history.json")
}

// savePinned writes only pinned items to disk as JSON.
func (a *App) savePinned() {
	a.mu.Lock()
	var pinned []ClipItem
	for _, item := range a.history {
		if item.Pinned {
			pinned = append(pinned, item)
		}
	}
	a.mu.Unlock()

	// Create directory if needed
	path := getHistoryFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		log.Printf("[clipboard] failed to create data directory: %v", err)
		return
	}

	// Write to file
	data, err := json.MarshalIndent(pinned, "", "  ")
	if err != nil {
		log.Printf("[clipboard] failed to marshal history: %v", err)
		return
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Printf("[clipboard] failed to write history file: %v", err)
	}
}

// loadHistory reads pinned items from disk on startup.
// All loaded items are marked as Pinned = true.
func (a *App) loadHistory() {
	path := getHistoryFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("[clipboard] failed to read history file: %v", err)
		}
		return
	}

	var pinned []ClipItem
	if err := json.Unmarshal(data, &pinned); err != nil {
		log.Printf("[clipboard] failed to unmarshal history: %v", err)
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Add pinned items to history (marked as pinned)
	for _, item := range pinned {
		if item.Text != "" {
			item.Pinned = true
			a.history = append(a.history, item)
		}
	}

	log.Printf("[clipboard] Loaded %d pinned items", len(pinned))
}

// TogglePin toggles the pinned state of an item at the given index.
// Exported for Wails binding.
func (a *App) TogglePin(index int) {
	a.mu.Lock()
	if index < 0 || index >= len(a.history) {
		a.mu.Unlock()
		log.Printf("[clipboard] TogglePin: invalid index %d", index)
		return
	}
	a.history[index].Pinned = !a.history[index].Pinned
	a.mu.Unlock()

	a.savePinned()
}

// DeleteItem removes an item from history at the given index.
// Exported for Wails binding.
func (a *App) DeleteItem(index int) {
	a.mu.Lock()
	if index < 0 || index >= len(a.history) {
		a.mu.Unlock()
		log.Printf("[clipboard] DeleteItem: invalid index %d", index)
		return
	}
	wasPinned := a.history[index].Pinned
	a.history = append(a.history[:index], a.history[index+1:]...)
	a.mu.Unlock()

	if wasPinned {
		a.savePinned()
	}
}
