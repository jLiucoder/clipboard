package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"golang.design/x/clipboard"
	"golang.org/x/image/draw"
)

// encodeBase64 encodes bytes to base64 string.
func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// decodeBase64 decodes base64 string, handling data URI prefix.
func decodeBase64(dataURI string) ([]byte, error) {
	// Remove data URI prefix if present
	if idx := strings.Index(dataURI, ","); idx != -1 {
		dataURI = dataURI[idx+1:]
	}
	return base64.StdEncoding.DecodeString(dataURI)
}

// resizeImage scales down large images to reduce memory usage.
// Max dimension is 300px (maintains aspect ratio).
func resizeImage(data []byte) ([]byte, error) {
	// Decode image
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		// If decode fails, return original (might be unsupported format)
		return data, nil
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// If already small enough, return as-is
	// Using 1200px max to keep text readable in screenshots (~30% file size reduction)
	maxDim := 1200
	if width <= maxDim && height <= maxDim {
		return data, nil
	}

	// Calculate new size maintaining aspect ratio
	var newWidth, newHeight int
	if width > height {
		newWidth = maxDim
		newHeight = int(float64(height) * float64(maxDim) / float64(width))
	} else {
		newHeight = maxDim
		newWidth = int(float64(width) * float64(maxDim) / float64(height))
	}

	// Create new image and scale
	newImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(newImg, newImg.Bounds(), img, bounds, draw.Over, nil)

	// Encode back to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, newImg); err != nil {
		return data, err
	}

	return buf.Bytes(), nil
}

// ClipItemType represents the type of clipboard content.
type ClipItemType string

const (
	TypeText  ClipItemType = "text"
	TypeImage ClipItemType = "image"
)

// ClipItem represents a single clipboard item.
type ClipItem struct {
	Type      ClipItemType `json:"type"`
	Text      string       `json:"text,omitempty"`
	ImageData string       `json:"imageData,omitempty"` // Base64 encoded image
	Pinned    bool         `json:"pinned"`
}

// initClipboard initializes the clipboard package.
func (a *App) initClipboard() error {
	return clipboard.Init()
}

// watchClipboard polls the clipboard for changes and captures content.
// Uses adaptive polling: faster when active, slower when idle to save CPU.
func (a *App) watchClipboard() {
	log.Println("[clipboard] Starting clipboard watcher...")

	// Initial change count
	lastCount := getPasteboardChangeCount()
	lastText := ""
	lastImageHash := ""
	idleTicks := 0

	for {
		// Adaptive sleep: 200ms when active, 1000ms when idle (5+ ticks no change)
		if idleTicks < 25 { // ~5 seconds of inactivity
			time.Sleep(200 * time.Millisecond)
		} else {
			time.Sleep(1000 * time.Millisecond) // Idle mode: check once per second
		}

		currentCount := getPasteboardChangeCount()
		if currentCount == lastCount {
			idleTicks++
			continue
		}

		// Activity detected, reset idle counter
		idleTicks = 0

		// Check if we recently pasted (within last 500ms) - skip to avoid capturing our own paste
		a.mu.Lock()
		timeSincePaste := time.Since(a.lastPasteTime)
		skipChange := currentCount == a.lastChangeCount || timeSincePaste < 500*time.Millisecond
		a.mu.Unlock()

		if skipChange {
			// This is our own paste, skip it
			lastCount = currentCount
			continue
		}

		// Try reading image first
		imgData := clipboard.Read(clipboard.FmtImage)
		if len(imgData) > 0 {
			// Simple hash check for duplicates
			hash := hashBytes(imgData)
			if hash != lastImageHash {
				lastImageHash = hash
				lastCount = currentCount
				a.addImageItem(imgData)
				log.Printf("[clipboard] Captured image (%d bytes)", len(imgData))
			} else {
				lastCount = currentCount
			}
			continue
		}

		// No image, try text
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

// hashBytes creates a simple hash for byte comparison.
func hashBytes(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	// Use first 16 bytes + length as simple hash
	n := 16
	if len(data) < n {
		n = len(data)
	}
	return string(data[:n]) + "_" + string(rune(len(data)))
}

// addItem adds a new text item to the clipboard history.
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
		if item.Type == TypeText && item.Text == text {
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
	newItem := ClipItem{Type: TypeText, Text: text, Pinned: false}
	a.history = append([]ClipItem{newItem}, a.history...)

	// Cap at 30 items, but preserve pinned items
	if len(a.history) > 30 {
		a.history = a.trimToCap()
	}
}

// addImageItem adds an image to the clipboard history.
// Images are resized to max 1200px to reduce memory usage while keeping text readable.
func (a *App) addImageItem(imgData []byte) {
	if len(imgData) == 0 {
		return
	}

	// Resize image to reduce memory (max 1200px dimension)
	resizedData, err := resizeImage(imgData)
	if err != nil {
		log.Printf("[clipboard] failed to resize image: %v", err)
		resizedData = imgData // Fallback to original
	}

	// Hash the RESIZED data for comparison (what we'll actually store)
	hash := hashBytes(resizedData)

	a.mu.Lock()
	defer a.mu.Unlock()

	// Skip if this image matches lastWritten (prevents re-capturing pasted images)
	if hash == hashBytes([]byte(a.lastWritten)) {
		a.lastWritten = ""
		return
	}

	// Encode resized image to base64 for storage
	imgBase64 := "data:image/png;base64," + encodeBase64(resizedData)

	// Check for duplicate images (compare by hash of resized data)
	for i, item := range a.history {
		if item.Type == TypeImage {
			// Decode stored image and hash it for comparison
			storedData, _ := decodeBase64(item.ImageData)
			if hashBytes(storedData) == hash {
				if item.Pinned {
					return
				}
				a.history = append(a.history[:i], a.history[i+1:]...)
				break
			}
		}
	}

	// Add new image item at the front
	newItem := ClipItem{Type: TypeImage, ImageData: imgBase64, Pinned: false}
	a.history = append([]ClipItem{newItem}, a.history...)

	// Log size reduction
	originalKB := len(imgData) / 1024
	resizedKB := len(resizedData) / 1024
	if originalKB > resizedKB {
		log.Printf("[clipboard] Image resized: %dKB → %dKB", originalKB, resizedKB)
	}

	// Cap at 30 items
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
	item := a.history[index]

	// Pre-compute lastWritten BEFORE writing to clipboard to avoid race condition
	var writeData []byte
	if item.Type == TypeImage {
		imgData, err := decodeBase64(item.ImageData)
		if err != nil {
			a.mu.Unlock()
			log.Printf("[clipboard] failed to decode image: %v", err)
			return
		}
		writeData = imgData
		a.lastWritten = hashBytes(imgData)
	} else {
		writeData = []byte(item.Text)
		a.lastWritten = item.Text
	}

	// Update change count and paste time BEFORE writing to clipboard
	a.lastChangeCount = getPasteboardChangeCount()
	a.lastPasteTime = time.Now()
	a.mu.Unlock()

	// Now write to clipboard (after lastWritten and lastChangeCount are set)
	if item.Type == TypeImage {
		clipboard.Write(clipboard.FmtImage, writeData)
	} else {
		clipboard.Write(clipboard.FmtText, writeData)
	}

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
