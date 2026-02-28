package main

import (
	"log"
	"os/exec"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// App is the Wails service that the frontend binds to.
type App struct {
	window   *application.WebviewWindow
	wailsApp *application.App

	prevAppPID string // PID of the app that was frontmost before we showed

	// Clipboard history
	mu              sync.Mutex
	history         []ClipItem
	lastChangeCount int
	lastWritten     string // Tracks text we just wrote to clipboard (to avoid re-capturing)
}

// capturePreviousApp records which app currently has focus so we can restore it later.
func (a *App) capturePreviousApp() {
	out, err := exec.Command("osascript", "-e",
		`tell application "System Events" to get unix id of first process whose frontmost is true`,
	).Output()
	if err != nil {
		log.Printf("[clipboard] capturePreviousApp failed: %v", err)
		a.prevAppPID = ""
		return
	}
	a.prevAppPID = strings.TrimSpace(string(out))
}

// restorePreviousApp re-activates the app that was focused before the island appeared.
func (a *App) restorePreviousApp() {
	if a.prevAppPID == "" {
		return
	}
	pid := a.prevAppPID
	a.prevAppPID = ""
	script := `tell application "System Events" to set frontmost of (first process whose unix id is ` + pid + `) to true`
	if err := exec.Command("osascript", "-e", script).Run(); err != nil {
		log.Printf("[clipboard] restorePreviousApp failed: %v", err)
	}
}

// HideWindow is called by the frontend on Escape — dismiss without pasting.
func (a *App) HideWindow() {
	a.window.Hide()
	go a.restorePreviousApp()
}
