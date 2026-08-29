# Agent notes

Miru is a Wails v2 desktop app (Go + React). The UI is a WebKitGTK webview, not a standalone Vite site.

## Checking the UI

After `make dev`, inspect the running app at **http://127.0.0.1:34115**.

That port is the Wails dev server with Go bindings. Do not use Vite `5173` or preview `4173` — those pages have no `window.go` / `window.runtime`, so React will not mount.

## Backend

Wails binds `App` in `package main`. Keep methods on `App`; split by concern into `app_*.go` (settings, library, anilist, playback, torrents). Do not add `internal/app` — that only relocates Wails glue and changes frontend imports (`wailsjs/go/main/App`).

Domain code lives in `internal/` (`storage`, `anilist`, `torrentx`, `mpv`, …). Split a file when it mixes responsibilities or passes ~700 lines. ~300–400 is comfortable; ~500 is fine for one type/flow.

Go: guard clauses over nested `if`. Invert the empty/error case and return. Extract a helper only if it is reused, or the nest is 3+ levels in the middle of a function. Do not flatten sequential checks (SQLite migrations, SOCKS5 client setup). Do not add a helper that is only called once. `make lint` is golangci-lint standard + nestif.

## Frontend

- App lives in `frontend/`. UI primitives are shadcn in `frontend/src/components/ui/`.
- Zero radius, 44px controls, purple only as the hit (Play, progress, selected mark). Use Tailwind tokens (`bg-background`, `text-muted-foreground`, `border-border`). Do not add CSS modules or new hex colors.
- Library poster grid and episode list stay custom. Do not flatten them into generic shadcn layouts.
- Views import icons from `frontend/src/components/Icons.tsx`, not Lucide directly.

Folders: `views/` one screen, `components/` app-owned (Sidebar, Icons), `components/ui/` generated shadcn (leave them), `lib/` pure helpers and DTOs. Do not add `hooks/`, `stores/`, or `features/` until something is reused across screens.

State stays local `useState` in the view. Shared tab/jobs/playback/toast live in `App.tsx` and pass as props. Wails `EventsOn` stays in `App`. Do not add Zustand, Redux, or a new Context. Do not edit `frontend/wailsjs/` (generated). Call bindings from `wailsjs/go/main/App`; DTO types from `lib/types.ts` — do not import `models.ts` in views.

Split a `.ts`/`.tsx` file when it mixes responsibilities or passes ~500 lines. Views: ~200–300 is comfortable; ~400 is fine for one screen. Extract a component or hook only if it is reused, or the file mixes distinct UI/logic. Logic with no JSX goes in `lib/`. Do not add a helper that is only called once.

Wails calls: `try/catch` and `errorMessage()`. Load failures: inline Alert. Action failures: `notice(..., true)`. Do not swallow errors.

Event handlers: keep a one-step `onClick`/`onChange` inline (`setState`, or `void namedFn()`). If the handler has a Wails call, `try/catch`, or more than one step, put it in a named function in the same view. Prefer `async`/`await` over `.then()` — including Wails loads in `useEffect` (`void reload()`, not `.then()`). After `await`, update state with `setForm((current) => …)`, not a closed-over `form`. Do not wrap in `useCallback`. Form `onSubmit` may stay inline for `preventDefault` plus a named save.

## Data location

Linux paths (from `internal/paths/paths.go`): `os.UserConfigDir()` + `miru`.

There is no separate config file. Settings live in the SQLite `settings` table.

- Config/DB: `$XDG_CONFIG_HOME/miru/app_data.db` (usually `~/.config/miru/app_data.db`)
- AniList token fallback (no Keyring): `~/.config/miru/anilist.token`
- Cache: `$XDG_CACHE_HOME/miru` (usually `~/.cache/miru`)
- Default downloads: `~/Downloads/miru`

`.env` in the repo root is build/dev only (AniList client ID/secret). It is not copied into `~/.config/miru/`.

## After editing

Always run the relevant checks before considering work done.

- Go changes: `make fmt` and `make lint`.
- Frontend changes: `make typecheck` and `make lint-fe`.
- Both: `make test` for Go.

Do not leave broken code — if a check fails, fix it before handing off.

## Verifying PRs

Verify pull requests yourself. Do not ask the user to run manual test-plan checkboxes when automated coverage already exists.

1. Fetch and checkout the PR branch (`gh pr checkout N` or `git fetch origin pull/N/head:pr-N && git checkout pr-N`).
2. Run the checks above for the changed areas. Shell commands that download Go modules or run golangci-lint may need unrestricted permissions (`required_permissions: ["all"]`) — the default sandbox can block the Go sumdb cache.
3. Read the PR diff and map each test-plan item to a unit test or an explicit gap. Prefer `go test -v -run 'Pattern' ./...` for targeted runs.
4. Report pass/fail with command output. Only mark something as “needs manual verification” when no test or static check covers it (e.g. visual UI polish, real torrent network behavior).

When the PR body lists manual steps, search `_test.go` for the behavior (status transitions, callbacks, redaction) before treating them as blockers.

## Git

Do not commit or push unless asked.
