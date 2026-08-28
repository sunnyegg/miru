package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"

	"github.com/sunnyegg/miru/internal/update"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) CheckForUpdate() (UpdateInfo, error) {
	info := UpdateInfo{Current: version}
	if update.IsDev(version) {
		return info, nil
	}
	if err := a.ready(); err != nil {
		return UpdateInfo{}, err
	}
	client, err := a.networkHTTPClient()
	if err != nil {
		return UpdateInfo{}, err
	}
	channel, err := a.updateChannel()
	if err != nil {
		return UpdateInfo{}, err
	}
	result, err := update.Check(a.ctx, client, version, update.ReleasesFeed, channel, goruntime.GOOS, goruntime.GOARCH)
	if err != nil {
		return UpdateInfo{}, err
	}
	return toUpdateInfo(result), nil
}

func (a *App) ApplyUpdate() error {
	if update.IsDev(version) {
		return errors.New("updates are disabled in development builds")
	}
	if err := a.ready(); err != nil {
		return err
	}
	client, err := a.networkHTTPClient()
	if err != nil {
		return err
	}
	downloadClient := *client
	downloadClient.Timeout = 0
	channel, err := a.updateChannel()
	if err != nil {
		return err
	}

	release, err := update.FetchLatest(a.ctx, client, update.ReleasesFeed, channel, goruntime.GOOS, goruntime.GOARCH)
	if err != nil {
		return err
	}
	if !update.Newer(version, release.Latest) {
		return errors.New("already up to date")
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	if err := update.Apply(a.ctx, &downloadClient, release.AssetURL, release.AssetName, exe); err != nil {
		return err
	}

	a.shutdown(a.ctx)
	if err := update.Restart(exe, os.Args, os.Environ()); err != nil {
		return fmt.Errorf("update installed but restart failed: %w; restart Miru manually", err)
	}
	if goruntime.GOOS == "windows" {
		runtime.Quit(a.ctx)
	}
	return nil
}

func (a *App) updateChannel() (string, error) {
	settings, err := a.loadSettings()
	if err != nil {
		return "", err
	}
	return settings.UpdateChannel, nil
}

func toUpdateInfo(result update.Info) UpdateInfo {
	return UpdateInfo{
		Current:    result.Current,
		Latest:     result.Latest,
		Available:  result.Available,
		Notes:      result.Notes,
		ReleaseURL: result.ReleaseURL,
		AssetName:  result.AssetName,
	}
}
