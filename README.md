# Miru

Miru is a portable desktop app for watching local anime in MPV, syncing
progress to AniList, and downloading torrents without a separate client.

Miru is currently in **alpha**. Expect rough edges and please report bugs on
the [issue tracker](https://github.com/sunnyegg/miru/issues).

## What's included

- **Library** — local files with AniList posters, Upcoming / Available sections for Watching and an On disk section filterable by AniList list status, each shelf grouped by airing season (Winter/Spring/Summer/Fall + year), a catch-up flow, and full AniList metadata (status, format, season, score, genres, studios, synopsis) with the next-episode air date and countdown on each show page
- **Playback** — MPV playback with per-episode resume and watch-position indicators that survive download cleanup, a now-playing bar with pause, seek, volume, full screen and stop controls, plus watch progress synced to AniList at a configurable threshold (default 85%)
- **AniList** — browser OAuth, list statuses, and an episode catalog on each show page
- **Profile** — AniList about text, title language, adult-content preference, and scoring format, plus your anime lists (Watching, Completed, Planning, and the rest); preview and save a banner PNG per list status for social media
- **Search** — AniList anime search by title or seasonal browse (season + year, popularity order, load more), plus Nyaa and Tokyo Toshokan torrent RSS with pagination
- **Anime details** — Library, Search, and Airing share the same complete metadata, including labeled genres and the next-episode countdown. Details share an account-scoped memory and SQLite cache for one hour, with cached data available when a refresh fails; search summaries and airing schedules remain separate.
- **Downloads** — built-in BitTorrent for magnet and `.torrent` links, with file selection, a concurrent limit, queue, bandwidth caps, and seeding across restarts
- **Airing** — weekly AniList schedule with controls for adding titles to Watching, Planning, or Completed; enabled after AniList responds successfully, with availability cached for six hours
- **Settings** — MPV path, native Wayland player window, Anime4K upscaling, Discord Rich Presence, close to system tray, download folder, speed limits, max concurrent downloads, seed ratio, RSS poll interval, auto-download from RSS, desktop notifications, torrent search title language and optional search prefix, network mode (system / direct / SOCKS5 / HTTP proxy), AniList, updates, and a full data reset
- **Updates** — Miru checks GitHub Releases and can update and restart from **Settings → About**

## Download

Download the latest build from [GitHub Releases](https://github.com/sunnyegg/miru/releases):

- `miru-<version>-linux-amd64` — Linux x86_64 standalone ELF
- `miru-<version>-windows-amd64.exe` — Windows x86_64 standalone with embedded WebView2
- `miru-<version>-mac-universal.zip` — macOS universal `.app` bundle for Apple Silicon and Intel

The macOS app is unsigned. On first launch, remove the quarantine attribute:

```bash
xattr -d com.apple.quarantine /Applications/miru.app
```

## Runtime dependencies (Linux)

- GTK3 and WebKitGTK
- MPV (`mpv` in `PATH`, or pick the binary in Settings)
- Optional: a Secret Service / GNOME Keyring. If missing, the AniList token is stored in `~/.config/miru/anilist.token` with mode `0600`.

On NVIDIA systems running Wayland, Miru automatically disables NVIDIA explicit
sync before starting WebKitGTK. This works around a known WebKitGTK protocol
error while keeping hardware acceleration enabled. If the window still crashes
or renders incorrectly, try the more conservative fallback:

```bash
WEBKIT_DISABLE_DMABUF_RENDERER=1 ./miru-*-linux-amd64
```

Debian/Ubuntu example:

```bash
sudo apt install libgtk-3-0 libwebkit2gtk-4.1-0 mpv
```

Package names differ by distro.

## First-run setup

1. Open **Settings**.
2. Confirm MPV is detected, or browse to a portable `mpv` binary.
3. Click **Open login** and authorize Miru in the browser.
   If the callback port is in use, paste the access token in Settings instead.

## Data location

- Config/DB: `$XDG_CONFIG_HOME/miru/app_data.db` (usually `~/.config/miru/`)
- MPV config: `$XDG_CONFIG_HOME/miru/mpv.conf` (included for Miru playback only)
- Cache: `$XDG_CACHE_HOME/miru`
- Logs: `$XDG_DATA_HOME/miru` (usually `~/.local/share/miru/`)
- Default downloads: `~/Downloads/miru`

Development builds (`make dev`, `go run`) use the same layout with a `-dev`
suffix: `~/.config/miru-dev`, `~/.cache/miru-dev`, `~/.local/share/miru-dev`,
`~/Downloads/miru-dev`, a `Miru Dev` window title, and their own single-instance
id. A local build can run next to an installed release without sharing settings,
database, login, logs, or downloads. `make clear-cache-dev` clears only the dev
cache.

## MPV window

The player window is titled `<app name> — <anime> — Episode <n>`, so dev builds
show `Miru Dev`. While an episode plays, the now-playing bar above the content
area controls pause, seek, volume, full screen and stop for the MPV window.

On Wayland, MPV runs through XWayland by default so the player window gets a
title bar, resize handles and a close button; GNOME gives native Wayland
clients none of those. Turn on **Settings → Playback → Use native Wayland for
the player window** to keep MPV on native Wayland instead.

The MPV window resizes freely instead of locking to the video aspect ratio, and
the OSC controls keep a readable size instead of shrinking with the window
(`script-opts=osc-vidscale=no`). Pin OSD text too with `osd-scale-by-window=no`,
or change the control size with `script-opts=osc-scalewindowed=<factor>`.
Miru passes two config files with `--include=`: a managed `mpv-defaults.conf`,
then your `mpv.conf`, then its own window flags. Options you add to `mpv.conf`
(`border=no`, `ontop=yes`, `autofit=70%x70%`, `keepaspect-window=yes`, and
similar) override the defaults. Miru writes `mpv.conf` with commented examples
on the first play and never overwrites your edits; `mpv-defaults.conf` is
rewritten on launch.

**Settings → About → Delete all data** removes Miru settings, library and
playback history, torrent history, RSS feeds, AniList login, cache, Anime4K
shaders, logs, and saved search state. Miru quits after confirmation and
finishes the reset the next time it opens. Downloaded media is not removed.
