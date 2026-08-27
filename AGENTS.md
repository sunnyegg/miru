# Agent notes

Miru is a Wails v2 desktop app (Go + React). The UI is a WebKitGTK webview, not a standalone Vite site.

## Checking the UI

After `make dev`, inspect the running app at **http://127.0.0.1:34115**.

That port is the Wails dev server with Go bindings. Do not use Vite `5173` or preview `4173` — those pages have no `window.go` / `window.runtime`, so React will not mount.

## Frontend

- App lives in `frontend/`. UI primitives are shadcn in `frontend/src/components/ui/`.
- Visual rules are in `DESIGN.md`: zero radius, 44px controls, orange only as the hit (Play, progress, selected mark).
- Library poster grid, OSC strip, and episode ticks stay custom. Do not flatten them into generic shadcn layouts.
- Views import icons from `frontend/src/components/Icons.tsx`, not Lucide directly.

## Data location

Linux paths (from `internal/paths/paths.go`): `os.UserConfigDir()` + `miru`.

There is no separate config file. Settings live in the SQLite `settings` table.

- Config/DB: `$XDG_CONFIG_HOME/miru/app_data.db` (usually `~/.config/miru/app_data.db`)
- AniList token fallback (no Keyring): `~/.config/miru/anilist.token`
- Cache: `$XDG_CACHE_HOME/miru` (usually `~/.cache/miru`)
- Default downloads: `~/Downloads/miru`

`.env` in the repo root is build/dev only (AniList client ID/secret). It is not copied into `~/.config/miru/`.

## Git

Do not commit or push unless asked.
