package main

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/sunnyegg/miru/internal/mpv"
	"github.com/sunnyegg/miru/internal/storage"
)

var errMissingDiscordAppID = errors.New("set DISCORD_APP_ID in .env or Settings")

func (a *App) discordAppID(settings SettingsView) string {
	if appID := strings.TrimSpace(settings.DiscordAppID); appID != "" {
		return appID
	}
	return envTrim("DISCORD_APP_ID")
}

func episodeAnimeTitle(episode storage.Episode) string {
	if episode.TitleEnglish != "" {
		return episode.TitleEnglish
	}
	if episode.TitleRomaji != "" {
		return episode.TitleRomaji
	}
	if episode.DisplayTitle != "" {
		return episode.DisplayTitle
	}
	return filepath.Base(episode.FilePath)
}

func (a *App) syncDiscordPresence(settings SettingsView, animeTitle string, episodeNumber int, percent float64) {
	if !settings.DiscordRpcEnabled {
		a.discord.Clear()
		return
	}
	appID := a.discordAppID(settings)
	if appID == "" {
		a.logDebugErr("discord rpc", errMissingDiscordAppID)
		return
	}
	if err := a.discord.Connect(appID); err != nil {
		a.logDebugErr("discord rpc connect", err)
		return
	}
	if err := a.discord.SetWatching(animeTitle, episodeNumber, percent); err != nil {
		a.logDebugErr("discord rpc update", err)
	}
}

func (a *App) clearDiscordPresence() {
	a.discord.Clear()
}

func (a *App) updateDiscordFromProgress(session *playSession, progress mpv.Progress, settings SettingsView) {
	if session == nil {
		return
	}
	a.syncDiscordPresence(settings, session.animeTitle, session.episodeNum, progress.Percent)
}
