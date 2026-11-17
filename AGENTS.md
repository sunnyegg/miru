# Agent notes

Miru is a Wails v2 desktop app (Go + React). The UI is a WebKitGTK webview, not a standalone Vite site.

## Checking the UI

After `make dev`, inspect the running app at **http://127.0.0.1:34115**.

`make dev-fresh` clears the dev cache first: it removes `$XDG_CACHE_HOME/miru-dev` and the dev SQLite `api_cache` table (AniList/detail/airing responses) via `scripts/clear-cache.sh miru-dev`, without touching settings, the AniList token, shaders, logs, or downloads. Use `make clear-cache` / `make clear-cache-dev` for the standalone step.

That port is the Wails dev server with Go bindings. Do not use Vite `5173` or preview `4173` — those pages have no `window.go` / `window.runtime`, so React will not mount.

## Backend

Wails binds `App` in `package main`. Keep methods on `App`; split by concern into `app_*.go` (settings, library, anilist, playback, torrents). Do not add `internal/app` — that only relocates Wails glue and changes frontend imports (`wailsjs/go/main/App`).

Domain code lives in `internal/` (`storage`, `anilist`, `torrentx`, `mpv`, …). Split a file when it mixes responsibilities or passes ~700 lines. ~300–400 is comfortable; ~500 is fine for one type/flow.

MPV control (pause, seek, volume, full screen) goes through `internal/mpv`'s `request`, which serializes IPC commands so two callers never read each other's reply. Pause, volume, full screen and stop stay in MPV's own OSC inside the player window; the now-playing strip (`components/NowPlayingStrip.tsx`) is a timeline only, calling `SeekPlayback` and storing the returned event in `playbackStore`, while the 1-second progress ticker emits the same event shape on `mpv:progress`. Seek commits when the slider is released or loses focus so a drag does not spam exact seeks. Keep the `App` method names and `PlaybackEvent` in step with `lib/types.ts` and the regenerated `wailsjs` bindings.

`PlayEpisode` hands mpv a playlist: the target episode plus the next downloaded episodes of the same show (`app_playback_queue.go`, contiguous, capped, stops at a gap). mpv advances by itself, so `followPlaylist` rotates the `playSession` when `mpv.Progress.Path` leaves the current entry: it saves and syncs the episode that ended, then swaps the episode fields, playback writer, window title and Discord presence, and seeks to the next episode's own resume position. `--start` is only passed for a single-file playlist because mpv applies it to every entry; a longer playlist seeks to the resume position after the file loads. `Player.media` keys its duration cache on `playlist-pos`, since each entry has its own length.

Go: guard clauses over nested `if`. Invert the empty/error case and return. Extract a helper only if it is reused, or the nest is 3+ levels in the middle of a function. Do not flatten sequential checks (SQLite migrations, SOCKS5 client setup). Do not add a helper that is only called once. `make lint` is golangci-lint standard + nestif.

## Frontend

- App lives in `frontend/`. UI primitives are shadcn in `frontend/src/components/ui/`.
- Zero radius, 44px controls, purple only as the hit (Play, progress, selected mark). Use Tailwind tokens (`bg-background`, `text-muted-foreground`, `border-border`). Do not add CSS modules or new hex colors.
- Library poster grid and episode list stay custom. Do not flatten them into generic shadcn layouts.
- Views import icons from `frontend/src/components/Icons.tsx`, not Lucide directly.

Folders: `views/` one screen, `components/` app-owned (Sidebar, Icons), `components/ui/` generated shadcn (leave them), `lib/` pure helpers and DTOs, `stores/` Zustand stores for cross-tab or persisted state. Do not add `hooks/` or `features/` until something is reused across screens.

State stays local `useState` in the view unless it must survive tab switches or persist — then use a Zustand store in `stores/` (one store per domain, e.g. `searchStore`). Domain state shared across tabs lives in stores; shell state (init, update banner, close prompt, toast) stays in `App.tsx`. Wails `EventsOn` stays in `App` and delegates to store actions via `getState()`. Do not add Redux or a new Context. Do not edit `frontend/wailsjs/` (generated). Call bindings from `wailsjs/go/main/App`; DTO types from `lib/types.ts` — do not import `models.ts` in views.

Split a `.ts`/`.tsx` file when it mixes responsibilities or passes ~500 lines. Views: ~200–300 is comfortable; ~400 is fine for one screen. Extract a component or hook only if it is reused, or the file mixes distinct UI/logic. Logic with no JSX goes in `lib/`. Do not add a helper that is only called once.

Wails calls: `try/catch` and `errorMessage()`. Load failures: inline Alert. Action failures: `notice(..., true)`. Do not swallow errors.

Event handlers: keep a one-step `onClick`/`onChange` inline (`setState`, or `void namedFn()`). If the handler has a Wails call, `try/catch`, or more than one step, put it in a named function in the same view. Prefer `async`/`await` over `.then()` — including Wails loads in `useEffect` (`void reload()`, not `.then()`). After `await`, update state with `setForm((current) => …)`, not a closed-over `form`. Do not wrap in `useCallback`. Form `onSubmit` may stay inline for `preventDefault` plus a named save.

## Data location

Linux paths (from `internal/paths/paths.go`): `os.UserConfigDir()` + `appinfo.Name`.

`internal/appinfo` is the single source of the app identity. A build with the `production` tag (added by `wails build`, including CI) is `miru` / `Miru` / `Miru.log` / `com.sunnyegg.miru`. Every other build (`wails dev`, `go run`, plain `go build`, tests) is the `-dev` variant: `miru-dev`, `Miru Dev`, `Miru-dev.log`, `com.sunnyegg.miru-dev`. A dev run therefore never shares config, cache, data, logs, the reset marker, the single-instance id, the window title, or the mpv window title with a release. Name-based checks and tests must use `appinfo` constants, not literals. Paths below show the release names; substitute `miru-dev` in a dev build.

There is no separate config file. Settings live in the SQLite `settings` table.

- Config/DB: `$XDG_CONFIG_HOME/miru/app_data.db` (usually `~/.config/miru/app_data.db`)
- AniList token: the platform keyring under the build's own service name (`appinfo.Name`: `miru` for a release, `miru-dev` for a dev build), so a dev build neither reads nor overwrites a release login; fallback file (no Keyring): `~/.config/miru/anilist.token`
- MPV config: `~/.config/miru/mpv-defaults.conf` (managed, rewritten on launch; currently `keepaspect-window=no` and `script-opts=osc-vidscale=no`) and `~/.config/miru/mpv.conf` (created once, never overwritten), both passed with `--include=` in that order so the user file overrides the defaults and Miru's own flags override both. `~/.config/miru/next.lua` (`internal/mpv/lua/next.lua`, embedded, rewritten on launch) is loaded with `--script=`; it draws the bottom-right **Next** button during an episode's last minute when the playlist has a following entry and advances with `playlist-next`. Its state is also exposed as `user-data/miru-next-visible`, `-hovered` and `-ass` for tests, and it only grabs `MBTN_LEFT` while the pointer is inside the button so mpv's own click behaviour stays intact
- Cache: `$XDG_CACHE_HOME/miru` (usually `~/.cache/miru`)
- Log: `$XDG_DATA_HOME/miru/Miru.log` (rotates to `Miru.log.1` at 5MB on startup; usually `~/.local/share/miru/Miru.log`)
- Default downloads: `~/Downloads/miru`
- Pending full reset marker: a `.miru-reset.json` sibling of the config `miru` directory (usually `~/.config/.miru-reset.json`); dev builds use `.miru-dev-reset.json`
- Staged reset backups: temporary `<miru-root>.reset-<id>` siblings of the config, cache, and data roots; config/data are deduplicated when they resolve to the same path

`.env` in the repo root is build/dev only (AniList client ID/secret). It is not copied into `~/.config/miru/`.

The Settings data reset quits Miru and processes the marker before the next startup opens the Wails logger or SQLite. It removes config, cache, logs, shaders, and known persisted frontend state, but never removes the configured download directory.

## After editing

Always run the relevant checks before considering work done.

- Go changes: `make fmt` and `make lint`.
- Frontend changes: `make fmt-fe`, `make typecheck`, and `make lint-fe`.
- Both: `make test` for Go.

Do not leave broken code — if a check fails, fix it before handing off.

Update documentation in the same change when behavior, settings, data paths, dependencies, or workflows change. `README.md` for user-facing features and setup; `AGENTS.md` for agent conventions and dev workflow; `docs/` for benchmarks and deep dives. Do not leave docs contradicting the code.

## Verifying PRs

Verify pull requests yourself. Do not ask the user to run manual test-plan checkboxes when automated coverage already exists.

1. Fetch and checkout the PR branch (`gh pr checkout N` or `git fetch origin pull/N/head:pr-N && git checkout pr-N`).
2. Run the checks above for the changed areas. Shell commands that download Go modules or run golangci-lint may need unrestricted permissions (`required_permissions: ["all"]`) — the default sandbox can block the Go sumdb cache.
3. Read the PR diff and map each test-plan item to a unit test or an explicit gap. Prefer `go test -v -run 'Pattern' ./...` for targeted runs.
4. Report pass/fail with command output. Only mark something as “needs manual verification” when no test or static check covers it (e.g. visual UI polish, real torrent network behavior).

When the PR body lists manual steps, search `_test.go` for the behavior (status transitions, callbacks, redaction) before treating them as blockers.

## Git

Do not commit or push unless asked.
