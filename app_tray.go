package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"github.com/gogpu/systray"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed assets/tray_icon.png
var trayIconPNG []byte

type closeAction int

const (
	closeActionAsk closeAction = iota
	closeActionHide
	closeActionQuit
)

func closeDecision(keyPresent, enabled, forceQuit bool) closeAction {
	if forceQuit {
		return closeActionQuit
	}
	if !keyPresent {
		return closeActionAsk
	}
	if enabled {
		return closeActionHide
	}
	return closeActionQuit
}

func (a *App) beforeClose(_ context.Context) (prevent bool) {
	if a.forceQuit.Load() {
		return false
	}

	keyPresent := false
	enabled := false
	if a.store != nil {
		keyPresent = !a.settingMissing("close_to_tray")
		if keyPresent {
			enabled = settingBool(a.store, "close_to_tray", false)
		}
	}

	switch closeDecision(keyPresent, enabled, false) {
	case closeActionQuit:
		return false
	case closeActionAsk:
		runtime.EventsEmit(a.ctx, "window:close-prompt")
		return true
	case closeActionHide:
		if a.hideToTray() {
			return true
		}
		return true
	default:
		return false
	}
}

func (a *App) ConfirmWindowClose(action string, remember bool) error {
	if err := a.ready(); err != nil {
		return err
	}

	normalizedAction := strings.ToLower(strings.TrimSpace(action))
	if normalizedAction != "hide" && normalizedAction != "quit" {
		return fmt.Errorf("unknown close action %q", action)
	}

	if remember {
		value := "false"
		if normalizedAction == "hide" {
			value = "true"
		}
		if err := a.store.SetSetting("close_to_tray", value); err != nil {
			return err
		}
	}

	if normalizedAction == "quit" {
		a.quitApp()
		return nil
	}

	if !a.hideToTray() {
		return errors.New("system tray is not available; Miru stayed open")
	}
	return nil
}

func (a *App) hideToTray() bool {
	if !a.trayReady.Load() {
		return false
	}
	runtime.WindowHide(a.ctx)
	return true
}

func (a *App) showWindow() {
	if a.ctx == nil {
		return
	}
	runtime.WindowShow(a.ctx)
	runtime.WindowUnminimise(a.ctx)
}

func (a *App) quitApp() {
	a.forceQuit.Store(true)
	runtime.Quit(a.ctx)
}

func (a *App) startTray() {
	a.trayMu.Lock()
	defer a.trayMu.Unlock()

	if a.tray != nil {
		return
	}

	tray := systray.New()
	menu := systray.NewMenu()
	menu.Add("Show", func() {
		a.showWindow()
	})
	menu.AddSeparator()
	menu.Add("Quit", func() {
		a.quitApp()
	})

	tray.SetIcon(trayIconPNG).
		SetTooltip("Miru").
		SetMenu(menu)
	tray.OnClick(func() {
		a.showWindow()
	})
	tray.Show()

	a.tray = tray
	a.trayReady.Store(true)

	go func() {
		if err := tray.Run(); err != nil {
			a.trayReady.Store(false)
			a.logDebugErr("system tray", err)
		}
	}()
}

func (a *App) stopTray() {
	a.trayMu.Lock()
	defer a.trayMu.Unlock()

	if a.tray == nil {
		return
	}
	a.tray.Remove()
	a.tray = nil
	a.trayReady.Store(false)
}
