package main

import (
	"errors"

	"github.com/sunnyegg/miru/internal/datareset"
	"github.com/sunnyegg/miru/internal/paths"
)

func (a *App) configureDataResetStartup(dirs paths.Dirs, state datareset.StartupState, err error) {
	a.dirs = dirs
	a.dataResetStartup = state
	a.dataResetStartupErr = err
	a.dataResetCleanupPending = state.CleanupPending
	if err != nil && !state.Blocked {
		a.dataResetError = redactError(err)
	}
}

func (a *App) finishDataResetStartup() {
	if !a.dataResetStartup.NeedsCommit {
		return
	}
	if err := datareset.Commit(a.dirs); err != nil {
		a.dataResetMu.Lock()
		a.dataResetCleanupPending = true
		a.dataResetError = redactError(err)
		a.dataResetMu.Unlock()
		a.logDebugErr("clean staged data reset", err)
		return
	}
	if a.dataResetStartupErr == nil {
		a.dataResetMu.Lock()
		a.dataResetCleanupPending = false
		a.dataResetError = ""
		a.dataResetMu.Unlock()
	}
}

func (a *App) GetDataSize() (DataSizeView, error) {
	if err := a.ready(); err != nil {
		return DataSizeView{}, err
	}
	usage, err := datareset.Measure(a.dirs)
	if err != nil {
		return DataSizeView{}, err
	}
	a.dataResetMu.RLock()
	cleanupPending := a.dataResetCleanupPending
	resetError := a.dataResetError
	a.dataResetMu.RUnlock()
	return DataSizeView{
		Bytes:          usage.Bytes,
		CleanupPending: usage.CleanupPending || cleanupPending,
		ResetError:     resetError,
	}, nil
}

func (a *App) DeleteAllData() error {
	if err := a.ready(); err != nil {
		return err
	}
	a.dataResetMu.Lock()
	defer a.dataResetMu.Unlock()
	if a.player != nil && a.player.Playing() {
		return errors.New("stop playback before deleting all Miru data")
	}
	if a.torrents != nil && a.torrents.Busy() {
		return errors.New("stop active downloads before deleting all Miru data")
	}
	if err := datareset.Schedule(a.dirs); err != nil {
		return err
	}
	a.forceQuit.Store(true)
	return nil
}
