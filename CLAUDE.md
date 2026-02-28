# Clipboard Island

macOS floating clipboard manager inspired by Windows' built-in clipboard. Built with Wails v3 (Go backend + vanilla JS frontend). Captures text and images, shows them in a floating panel at your cursor.

## Features

- **Text & Image Support** - Copy text or screenshots (Cmd+Shift+4), both appear in the panel
- **30 Item History** - Automatically capped, pinned items preserved
- **Pin Items** - ☆/★ toggle to keep items across restarts
- **One-Click Paste** - Click or press Enter to paste at cursor position
- **Keyboard Navigation** - ↑/↓ arrows to select, Enter to paste, Escape to dismiss
- **Persistent Pins** - Pinned items saved to `~/.clipboard-island/history.json`
- **Image Resizing** - Screenshots resized to 1200px max to save memory while keeping text readable
- **Duplicate Prevention** - Same content won't be added twice; pasting won't re-capture

## Build & Run

```bash
wails3 build                    # produces bin/clipboard-island binary
./package.sh                    # creates bin/Clipboard Island.app bundle
go test ./...                   # run unit tests (46 tests)
open 'bin/Clipboard Island.app' # run the app (no terminal, no dock icon)
```

**Important**: Don't run `./bin/clipboard-island` directly - it will show a terminal window! Always use the .app bundle or `open` command.

Frontend is built automatically by `wails3 build`. To rebuild frontend only: `cd frontend && npm run build`.

## Usage

1. **Copy** text or image (Cmd+C or screenshot)
2. **Open** clipboard island with Cmd+Shift+V
3. **Navigate** with arrow keys or mouse hover
4. **Paste** selected item with Enter or click
5. **Pin** important items with ☆ button (persists across restarts)
6. **Delete** items with × button

## Architecture

| Layer | Tech |
|-------|------|
| Backend | Go, Wails v3 alpha.73 |
| Frontend | Vanilla JS + Vite, `@wailsio/runtime` |
| Hotkey | `golang.design/x/hotkey` (Cmd+Shift+V to show) |
| Clipboard | `golang.design/x/clipboard` (text + image support) |
| Window chrome | Frameless, transparent background, CGo for macOS APIs |
| Persistence | `github.com/adrg/xdg` (XDG data directory) |
| Icon | Custom generated (black bg, white clipboard, fits macOS dark theme) |

### Key files

- `main.go` — Wails app bootstrap, window config, tray icon, Cmd+Shift+V hotkey goroutine, clipboard watching
- `app.go` — `App` service struct, focus capture/restore (`capturePreviousApp`, `restorePreviousApp`), `HideWindow()`, `lastPasteTime` tracking
- `clipboard.go` — Core clipboard logic: `ClipItem` struct, `addItem()`, `addImageItem()`, `GetHistory()`, `TogglePin()`, `DeleteItem()`, `SelectItem()`, persistence
- `clipboard_darwin.go` — CGo helper for `NSPasteboard changeCount`
- `clipboard_test.go` — 46 unit tests covering text, images, pins, caps, dedup
- `position.go` — Pure `calcWindowPosition()` function (flip-anchor logic to keep panel in-bounds)
- `position_test.go` — Unit tests for positioning: edge cases, boundary sweep
- `cursor_darwin.go` — CGo helper: mouse position + screen size in scaled pixels (macOS only)
- `frontend/src/main.js` — Renders history list, handles hotkey/keyboard/click events, keyboard navigation
- `frontend/public/style.css` — macOS-native dark blur panel, header styles, clip rows, pin/delete buttons, selected state
- `frontend/index.html` — Panel shell with draggable header and scrollable body
- `package.sh` — Creates .app bundle (run after `wails3 build` to get a terminal-free app)
- `generate_icon.go` — Go script to generate the app icon (black background, white clipboard)
- `update_icon.sh` — Converts icon_1024.png to .icns format
- `icon_1024.png` — Source icon file (1024x1024)

### Coordinate system (important)

All positioning uses **scaled pixels** (points * backingScaleFactor). `cursor_darwin.go` returns cursor and screen dimensions in this space. Wails' `SetPosition()` expects scaled pixels — it divides by scale internally. Never mix point-space values with scaled-pixel values.

### Window sizing

Window width is locked at 380px (MinWidth = MaxWidth). Height is 370px. User can drag-resize vertically (min 120, max 800). The CSS uses flex layout so the body fills available space.

### Focus capture

On hotkey, `capturePreviousApp()` records the frontmost app's PID via AppleScript. On dismiss (Escape or click-outside), `restorePreviousApp()` re-activates it. When pasting, the app:
1. Writes content to clipboard
2. Hides the window
3. Restores previous app focus
4. Simulates Cmd+V keystroke

### Clipboard watching

The watcher polls every 200ms (1s when idle) and:
1. Checks `changeCount` to detect changes
2. Tries to read image first, then text
3. Skips if within 500ms of a paste (prevents self-capture)
4. Resizes images to max 1200px before storing
5. Stores as base64 data URLs for frontend display

### Data model

```go
type ClipItem struct {
    Type      string  // "text" or "image"
    Text      string  // For text items
    ImageData string  // Base64 data URL for images
    Pinned    bool    // Persisted to disk
}
```

History is capped at 30 items. Pinned items are never evicted. On startup, pinned items are loaded from `~/.clipboard-island/history.json`.

## Platform notes

- macOS only (CGo, AppleScript, `ActivationPolicyAccessory`)
- `/usr/sbin` is added to PATH at startup so Wails can find `sysctl`
- Linker warnings about macOS version mismatch are harmless (Go CGo on newer macOS)
- The `--wails-draggable: drag` CSS property on `#island-header` enables window dragging
- `HideOnFocusLost: true` handles click-outside dismiss automatically

## Module name

The Go module is named `changeme` (Wails scaffold default). Bindings are generated into `frontend/bindings/changeme/`.
