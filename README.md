# Miru

Portable Linux desktop app for playing local anime files in MPV, syncing watch progress to AniList, and downloading one torrent at a time.

## Runtime dependencies (Linux)

- GTK3 and WebKitGTK (Wails webview)
- MPV (`mpv` in `PATH`, or pick the binary in Settings)
- Optional: a Secret Service / GNOME Keyring. If missing, the AniList token is stored in `~/.config/miru/anilist.token` with mode `0600`.

On NVIDIA systems running Wayland, Miru automatically disables NVIDIA explicit
sync before starting WebKitGTK. This works around a known WebKitGTK protocol
error while keeping hardware acceleration enabled. If the window still crashes
or renders incorrectly, try the more conservative fallback:

```bash
WEBKIT_DISABLE_DMABUF_RENDERER=1 ./build/bin/miru-linux-amd64
```

Debian/Ubuntu example:

```bash
sudo apt install libgtk-3-0 libwebkit2gtk-4.1-0 mpv
```

Package names differ by distro. If `wails doctor` reports a missing webkit package, install the one it names.

## Build

Requires Go 1.22+, Node.js, and [Wails v2](https://wails.io).

```bash
make doctor
make build
```

Pop!_OS / Ubuntu 24.04 ships WebKitGTK 4.1. The `webkit2_41` tag is required. The Linux binary is written to `build/bin/miru-linux-amd64`.

The release build uses Wails' `production` build tag and embeds the local `.env`
file into the binary. Set the AniList values in `.env` before running
`wails build`; the resulting binary does not need a `.env` file beside it.

Use `make help` to see all available commands. `make dev` runs development
mode, while `make test`, `make fmt`, and `make clean` run tests, format Go
source, and remove generated output.

## First-run setup

1. Copy `.env.example` to `.env` and set `ANILIST_CLIENT_ID` and `ANILIST_CLIENT_SECRET`.
   Keep `.env` local; it is ignored by Git.
2. Open **Settings**.
3. Confirm MPV is detected, or browse to a portable `mpv` binary.
4. Click **Open login** and authorize in the browser. Miru exchanges the code at `http://127.0.0.1:58496/callback`.
   - Redirect URL on the AniList app must be `http://127.0.0.1:58496/callback` (click Save on that form).
   - If the port is in use, paste the access token in Settings instead.

## MVP scope

Included: local library, MPV playback, AniList progress at ≥85%, one magnet/`.torrent` download that stops uploading when complete.

Deferred: RSS indexers, airing calendar, bandwidth sliders, 0.5× seeding ratio, Windows/macOS packages.

## Smoke notes (Linux)

- `go test ./...` passes.
- `wails build -tags webkit2_41` produces `build/bin/miru-linux-amd64` (~27 MB).
- App window starts on a local display. Idle RAM for the process tree (app + WebKit) was about 400 MB on Pop!_OS 24.04; the PRD 100 MB idle target is a later optimization, not an MVP blocker.
- MPV is not bundled. Install `mpv` or point Settings at a portable binary before Play will work.

## Data location

- Config/DB: `$XDG_CONFIG_HOME/miru/app_data.db` (usually `~/.config/miru/`)
- Cache: `$XDG_CACHE_HOME/miru`
- Default downloads: `~/Downloads/miru`
