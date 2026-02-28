package main

import (
	"embed"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/icons"
	"golang.design/x/hotkey"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Ensure /usr/sbin is in PATH so Wails can find sysctl.
	if !strings.Contains(os.Getenv("PATH"), "/usr/sbin") {
		os.Setenv("PATH", os.Getenv("PATH")+":/usr/sbin")
	}

	appService := &App{}

	wailsApp := application.New(application.Options{
		Name:        "Clipboard",
		Description: "Clipboard manager",
		Services: []application.Service{
			application.NewService(appService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ActivationPolicy:                                 application.ActivationPolicyAccessory,
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	// ── Island window ─────────────────────────────────────────────────────────
	const windowW = 380
	const windowH = 370

	window := wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:           "Clipboard",
		Width:           windowW,
		Height:          windowH,
		MinWidth:        windowW,
		MaxWidth:        windowW,
		MinHeight:       120,
		MaxHeight:       800,
		Frameless:       true,
		AlwaysOnTop:     true,
		Hidden:          true,
		HideOnFocusLost: true,
		BackgroundType:  application.BackgroundTypeTransparent,
		Mac: application.MacWindow{
			Backdrop: application.MacBackdropTransparent,
		},
		URL: "/",
	})

	appService.window = window
	appService.wailsApp = wailsApp

	// Helper to show the island (used by both hotkey and tray menu)
	showIsland := func() {
		appService.capturePreviousApp()
		cx, cy, sw, sh, scale := cursorAndScreen()
		ww := windowW * scale
		wh := windowH * scale
		wx, wy := calcWindowPosition(cx, cy, ww, wh, sw, sh)
		window.SetSize(windowW, windowH)
		window.SetPosition(wx, wy)
		window.Show()
		window.Focus()
		wailsApp.Event.Emit("hotkey")
	}

	// ── Menu bar icon ─────────────────────────────────────────────────────────
	tray := wailsApp.SystemTray.New()
	if runtime.GOOS == "darwin" {
		tray.SetTemplateIcon(icons.SystrayMacTemplate)
	}
	trayMenu := wailsApp.NewMenu()
	trayMenu.Add("Show Clipboard  (\u2318\u21e7V)").OnClick(func(ctx *application.Context) {
		showIsland()
	})
	trayMenu.AddSeparator()
	trayMenu.Add("Quit Clipboard").OnClick(func(ctx *application.Context) {
		wailsApp.Quit()
	})
	tray.SetMenu(trayMenu)
	tray.SetTooltip("Clipboard")

	// ── Global hotkey ─────────────────────────────────────────────────────────
	go func() {
		hk := hotkey.New([]hotkey.Modifier{hotkey.ModCmd, hotkey.ModShift}, hotkey.KeyV)
		if err := hk.Register(); err != nil {
			log.Printf("[clipboard] hotkey register failed (grant Accessibility): %v", err)
			return
		}
		log.Println("[clipboard] Hotkey Cmd+Shift+V active")
		for range hk.Keydown() {
			showIsland()
		}
	}()

	if err := wailsApp.Run(); err != nil {
		log.Fatal(err)
	}
}
