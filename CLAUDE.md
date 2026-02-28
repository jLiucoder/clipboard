# Clipboard

macOS floating panel app. Built with Wails v3 (Go backend + vanilla JS frontend). Currently a minimal shell — hotkey shows the window, Escape dismisses it.

## Build & Run

```bash
wails3 build                    # produces bin/clipboard-island
go test ./...                   # run unit tests
./bin/clipboard-island          # run the app (tray icon, no dock icon)
```

Frontend is built automatically by `wails3 build`. To rebuild frontend only: `cd frontend && npm run build`.

**Important**: `go run .` only compiles Go — it embeds whatever is in `frontend/dist`. If you changed HTML/CSS/JS, you must run `wails3 build` (or `cd frontend && npm run build` first).

## Architecture

| Layer | Tech |
|-------|------|
| Backend | Go, Wails v3 alpha.73 |
| Frontend | Vanilla JS + Vite, `@wailsio/runtime` |
| Hotkey | `golang.design/x/hotkey` (Cmd+Shift+V to show) |
| Window chrome | Frameless, transparent background, CGo for macOS APIs |

### Key files

- `main.go` — Wails app bootstrap, window config, tray icon, Cmd+Shift+V hotkey goroutine, `showIsland()`
- `app.go` — `App` service struct, focus capture/restore (`capturePreviousApp`, `restorePreviousApp`), `HideWindow()`
- `position.go` — Pure `calcWindowPosition()` function (flip-anchor logic to keep panel in-bounds)
- `position_test.go` — Unit tests for positioning: edge cases, boundary sweep
- `cursor_darwin.go` — CGo helper: mouse position + screen size in scaled pixels (macOS only)
- `frontend/src/main.js` — Listens for `hotkey` event to show island, Escape to dismiss
- `frontend/public/style.css` — macOS-native dark blur panel, header styles
- `frontend/index.html` — Panel shell with draggable header and empty body

### Coordinate system (important)

All positioning uses **scaled pixels** (points * backingScaleFactor). `cursor_darwin.go` returns cursor and screen dimensions in this space. Wails' `SetPosition()` expects scaled pixels — it divides by scale internally. Never mix point-space values with scaled-pixel values.

### Window sizing

Window width is locked at 380px (MinWidth = MaxWidth). Height is 370px. User can drag-resize vertically (min 120, max 800). The CSS uses flex layout so the body fills available space.

### Focus capture

On hotkey, `capturePreviousApp()` records the frontmost app's PID via AppleScript. On dismiss (Escape or click-outside), `restorePreviousApp()` re-activates it.

## Platform notes

- macOS only (CGo, AppleScript, `ActivationPolicyAccessory`)
- `/usr/sbin` is added to PATH at startup so Wails can find `sysctl`
- Linker warnings about macOS version mismatch are harmless (Go CGo on newer macOS)
- The `--wails-draggable: drag` CSS property on `#island-header` enables window dragging
- `HideOnFocusLost: true` handles click-outside dismiss automatically

## Module name

The Go module is named `changeme` (Wails scaffold default). Bindings are generated into `frontend/bindings/changeme/`.
