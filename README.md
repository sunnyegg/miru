# Miru

Miru is a portable desktop app for watching local anime in MPV, syncing
progress to AniList, and downloading torrents without a separate client.

Miru is currently in **alpha**. Expect rough edges and please report bugs on
the [issue tracker](https://github.com/sunnyegg/miru/issues).

## What's included

- **Library** — local files with AniList posters, Watching / Unlisted sections, and a catch-up flow
- **Playback** — MPV playback with watch progress synced to AniList at a configurable threshold (default 85%)
- **AniList** — browser OAuth, list statuses, and an episode catalog on each show page
- **Search** — Nyaa and Tokyo Toshokan RSS with pagination
- **Downloads** — built-in BitTorrent for magnet and `.torrent` links, with file selection, a concurrent limit, queue, bandwidth caps, and seeding across restarts
- **Airing** — weekly schedule of upcoming episodes from AniList
- **Settings** — MPV path, Anime4K upscaling, Discord Rich Presence, close to system tray, download folder, speed limits, max concurrent downloads, seed ratio, RSS poll interval, auto-download from RSS, desktop notifications, network mode (system / direct / SOCKS5 / HTTP proxy), AniList, and About
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
- Cache: `$XDG_CACHE_HOME/miru`
- Default downloads: `~/Downloads/miru`
