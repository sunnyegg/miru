package main

import (
	"embed"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	applyWaylandNVIDIAWorkaround()

	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "Miru",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 15, B: 35, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		OnBeforeClose:    app.beforeClose,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "com.sunnyegg.miru",
			OnSecondInstanceLaunch: func(_ options.SecondInstanceData) {
				app.showWindow()
			},
		},
		Bind: []interface{}{
			app,
		},
		Linux: &linux.Options{
			WindowIsTranslucent: false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func applyWaylandNVIDIAWorkaround() {
	if os.Getenv("WAYLAND_DISPLAY") == "" {
		return
	}
	if _, err := os.Stat("/sys/module/nvidia"); err != nil {
		return
	}
	if os.Getenv("__NV_DISABLE_EXPLICIT_SYNC") == "" {
		_ = os.Setenv("__NV_DISABLE_EXPLICIT_SYNC", "1")
	}
}
