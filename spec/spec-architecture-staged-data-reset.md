---
title: Miru Staged Application Data Reset
version: 1.0
date_created: 2026-09-02
last_updated: 2026-09-02
owner: Miru maintainers
tags: [architecture, data, reset, settings, wails]
---

# Introduction

This specification defines a safe, user-initiated reset of Miru application data. The reset removes Miru-managed configuration, state, cache, logs, Anime4K shaders, authentication data, and known persisted frontend state. Downloaded media remains untouched.

Miru must perform filesystem deletion after a normal application shutdown. The user confirms the reset in Settings, Miru exits, and the reset is completed when the user opens Miru again. Automatic relaunch is outside this specification.

## 1. Purpose & Scope

This specification covers GitHub issue [#44](https://github.com/sunnyegg/miru/issues/44).

The intended audience is an implementation agent working on the Wails v2 Go backend and React frontend. The implementation must work on Linux, macOS, and Windows.

In scope:

- Calculate the filesystem size of Miru-managed configuration, cache, data, and staged reset backups.
- Present a destructive reset action in Settings > About.
- Require a confirmation dialog before scheduling the reset.
- Disable the action while MPV playback is active.
- Stop torrent transfers, RSS polling, AniList login callbacks, Discord presence, and other application work through the normal shutdown path.
- Stage the reset by renaming data roots before creating a fresh application state.
- Preserve recoverability when staging fails.
- Report incomplete cleanup instead of silently claiming success.
- Remove known Miru `localStorage` entries used by Zustand persistence.

Out of scope:

- Files under the configured download directory, including the default `~/Downloads/miru` directory.
- Secure erasure guarantees for solid-state storage, snapshots, backups, or filesystem journals.
- Automatic process relaunch.
- A selective reset that preserves settings or history.
- Removal of WebKitGTK engine data that is not stored in Miru-managed paths and is not represented by a known Miru `localStorage` key.

## 2. Definitions

- **Active root**: A directory currently used by Miru.
- **Config root**: `paths.Dirs.Config`. It contains `app_data.db`, SQLite sidecar files, `anilist.token`, and Anime4K shaders.
- **Cache root**: `paths.Dirs.Cache`.
- **Data root**: `paths.Dirs.Data`. It contains `Miru.log` and rotated logs.
- **Downloaded media**: User content saved to the configured download directory. It is not application state.
- **Marker**: A small JSON file stored beside, not inside, the config root. It records reset progress and backup names.
- **Reset ID**: A cryptographically random identifier used to create collision-resistant backup names.
- **Staging**: Renaming each active root to a sibling backup path on the same filesystem.
- **Commit**: Creating fresh active roots and successfully initializing a new Miru database.
- **Cleanup**: Permanently removing staged backup roots and the marker.
- **Rollback**: Renaming staged backup roots back to their active paths after a staging failure.
- **Fresh state**: A newly initialized Miru installation with no prior settings, library history, feeds, torrent history, AniList token, shaders, cache, logs, or known persisted frontend state.
- **Playback active**: `mpv.Player.Playing()` returns `true`, or the frontend playback store contains a current playback event.

## 3. Requirements, Constraints & Guidelines

### Functional requirements

- **REQ-001**: Settings > About shall show the current size of Miru-managed filesystem data.
- **REQ-002**: Size calculation shall include all entries under the config, cache, and data roots.
- **REQ-003**: Size calculation shall include reset backup roots referenced by an existing marker.
- **REQ-004**: Size calculation shall return raw bytes. The frontend shall format bytes using the existing `formatBytes` helper.
- **REQ-005**: The reset action shall use a destructive button variant and shall state that the action cannot be undone.
- **REQ-006**: The reset flow shall require two user actions: open the confirmation dialog, then activate the destructive confirmation button.
- **REQ-007**: The confirmation dialog shall state that settings, history, AniList login, RSS feeds, torrent history, cache, shaders, logs, and persisted search/playback UI state will be removed.
- **REQ-008**: The confirmation dialog shall state that downloaded media will remain on disk.
- **REQ-009**: The confirmation dialog shall state that active torrent transfers will stop and that the user must open Miru again to finish the reset.
- **REQ-010**: The reset action shall be disabled while playback is active.
- **REQ-011**: The backend shall independently reject reset scheduling while playback is active. Frontend state is not an authority for destructive operations.
- **REQ-012**: Active torrent transfers shall not block reset scheduling. The normal application shutdown path shall stop them before reset staging.
- **REQ-013**: Reset scheduling shall create the marker atomically before the frontend quits Miru. Later marker transitions shall replace it atomically.
- **REQ-014**: After reset scheduling succeeds, the frontend shall remove the `miru.playback` and `miru.search` `localStorage` entries and invoke the Wails runtime quit operation.
- **REQ-015**: If scheduling fails, Miru shall remain open and show the error inside the destructive section or confirmation dialog.
- **REQ-016**: Pending reset processing shall run before Wails creates its file logger and before `storage.Open` opens SQLite.
- **REQ-017**: Pending reset processing shall stage existing active roots by renaming them to unique sibling backup paths.
- **REQ-018**: The marker shall record enough state to make every staging and cleanup step idempotent after a crash.
- **REQ-019**: If staging fails, Miru shall roll back every root that was staged during the attempt.
- **REQ-020**: If staging rollback succeeds, Miru shall start with the previous data and expose a reset failure message.
- **REQ-021**: If staging rollback cannot restore a consistent set of active roots, Miru shall not initialize storage against the partial state. It shall expose a blocking initialization error and preserve the marker and backups for another attempt.
- **REQ-022**: After staging succeeds, Miru shall recreate the active roots with the permissions used by `paths.Resolve` and initialize a fresh SQLite database through `storage.Open`.
- **REQ-023**: Miru shall not delete reset backups until the fresh database has initialized successfully.
- **REQ-024**: After fresh initialization succeeds, Miru shall remove staged backup roots and then remove the marker.
- **REQ-025**: If cleanup fails, Miru shall keep the marker, retain a fresh active state, report that old data remains staged, and retry cleanup on a later startup.
- **REQ-026**: A cleanup failure shall not restore old settings into an already initialized fresh database.
- **REQ-027**: After a complete reset, `GetDataSize` shall report the size of the newly initialized files. The value is not required to be literal `0 B` because SQLite and the logger create new files during startup.
- **REQ-028**: Roots that resolve to the same cleaned absolute path shall be measured, staged, rolled back, and cleaned exactly once. This applies to config and data roots on platforms where both use the same application-data directory.

### Security and safety requirements

- **SEC-001**: Reset operations shall only target paths derived from `paths.Dirs` and backup paths derived from those exact roots plus the current Reset ID.
- **SEC-002**: The implementation shall reject an empty root, filesystem root, home directory, root parent directory, or path whose base name is not `miru`.
- **SEC-003**: Recursive size calculation shall use link metadata and shall not follow symbolic links.
- **SEC-004**: Staging and cleanup shall not follow symbolic links outside a reset root.
- **SEC-005**: The marker shall be written with owner-only permissions where the platform supports Unix file modes.
- **SEC-006**: Marker writes shall use a temporary sibling file followed by rename so a crash cannot leave partially written JSON as the only reset record.
- **SEC-007**: Error messages shown to users shall not expose AniList tokens, torrent URLs, or unrelated absolute paths.
- **SEC-008**: A second reset request shall be rejected while a reset marker already represents pending or incomplete work.

### Constraints and guidelines

- **CON-001**: The implementation shall use the Go standard library and existing project dependencies. No new dependency is required.
- **CON-002**: The reset workflow must support Linux, macOS, and Windows file-lock behavior.
- **CON-003**: The Wails logger owns the current log file. Therefore, data-root staging must occur in a later process before the new logger is created.
- **CON-004**: A multi-root filesystem operation cannot be globally atomic when XDG roots reside on different filesystems. The marker and rollback rules provide recoverable, idempotent behavior instead.
- **CON-005**: Existing generated files under `frontend/wailsjs/` shall not be edited manually.
- **GUD-001**: Reuse the existing zero-radius card, dialog, button, spacing, and semantic color patterns.
- **GUD-002**: The destructive confirmation button shall not receive initial focus. Initial focus shall prefer Cancel.
- **GUD-003**: The dialog shall close with Escape before confirmation and shall not close while scheduling is in progress.
- **GUD-004**: All actionable controls shall retain visible keyboard focus and a minimum 44 by 44 pixel target.
- **GUD-005**: The UI shall not use color as the only indication that the action is destructive.
- **GUD-006**: Double submission shall be prevented in both frontend and backend state.
- **PAT-001**: App-bound methods shall remain on `App`; testable filesystem coordination shall live in one focused package under `internal/`.

## 4. Interfaces & Data Contracts

### 4.1 Wails App methods

```go
type DataSizeView struct {
    Bytes          int64  `json:"bytes"`
    CleanupPending bool   `json:"cleanupPending"`
    ResetError     string `json:"resetError"`
}

func (a *App) GetDataSize() (DataSizeView, error)
func (a *App) DeleteAllData() error
```

`GetDataSize` behavior:

- Returns active-root size plus backup-root size referenced by the marker.
- Sets `CleanupPending` when staged backup data still exists.
- Sets `ResetError` when the current process received a recoverable staging or cleanup error.
- Returns an error when the size cannot be calculated reliably.

`DeleteAllData` behavior:

- Checks application readiness.
- Rejects the call if MPV playback is active.
- Rejects concurrent or duplicate reset scheduling.
- Creates the marker atomically.
- Sets the existing force-quit state so the close-to-tray prompt cannot intercept the frontend quit call.
- Returns only after the marker is durable.
- Does not delete files or call `runtime.Quit` itself.

The frontend owns the following ordered sequence after `DeleteAllData` succeeds:

```ts
localStorage.removeItem('miru.playback')
localStorage.removeItem('miru.search')
Quit()
```

### 4.2 Reset marker

The marker shall be a sibling of the config root and shall use a stable filename such as `.miru-reset.json`.

```json
{
  "schemaVersion": 1,
  "resetId": "random-lowercase-hex",
  "phase": "pending",
  "stagedRoots": ["config"]
}
```

Allowed phases:

- `pending`: The request is durable and staging is incomplete.
- `staged`: All prior roots are backups and fresh roots may be created.
- `cleanup_pending`: Fresh initialization succeeded, but one or more backups remain.

`stagedRoots` may contain only `config`, `cache`, and `data`. Backup paths shall be derived as `<active-root>.reset-<resetId>` and shall not be accepted from marker input.

### 4.3 Startup integration

`main.go` shall resolve Miru paths, process reset staging or pending cleanup, and only then construct the Wails file logger. The result shall be passed to `App` as in-memory reset status. After `App.init` successfully initializes fresh state, startup shall request backup cleanup.

### 4.4 Frontend component contract

The About panel shall receive:

```ts
type DataSizeView = {
  bytes: number
  cleanupPending: boolean
  resetError: string
}

type DataResetProps = {
  dataSize: DataSizeView | null
  dataSizeError: string
  playbackActive: boolean
  resetting: boolean
  onReloadDataSize: () => void
  onDeleteAllData: () => void
}
```

The exact prop split may follow existing Settings conventions, but the observable states and actions shall remain equivalent.

## 5. Acceptance Criteria

- **AC-001**: Given Miru contains a database, token, shaders, cache, and logs, when Settings > About loads, then it displays their combined filesystem size in a human-readable format.
- **AC-002**: Given no Miru files exist, when size is loaded, then the UI displays `0 B`.
- **AC-003**: Given MPV playback is active, when the About panel is shown, then the reset button is disabled and explains why.
- **AC-004**: Given stale frontend playback state incorrectly reports no playback, when the backend detects MPV is active, then `DeleteAllData` rejects the request without creating a marker.
- **AC-005**: Given playback is inactive, when the user activates the reset button, then a keyboard-accessible confirmation dialog appears and Cancel has safe initial focus.
- **AC-006**: Given the dialog is open, when the user cancels or presses Escape, then no marker or data change occurs.
- **AC-007**: Given the user confirms, when marker creation succeeds, then known Miru `localStorage` keys are removed and Miru quits normally.
- **AC-008**: Given active torrent transfers exist, when reset is confirmed, then shutdown stops them and downloaded media remains on disk.
- **AC-009**: Given a pending marker, when Miru starts again, then root staging occurs before the logger and SQLite database are opened.
- **AC-010**: Given all roots stage successfully, when fresh initialization completes, then prior settings, history, feeds, torrent jobs, AniList authentication, shaders, cache, and logs are absent from active roots.
- **AC-011**: Given staging fails after one root was renamed, when rollback succeeds, then all staged roots return to their original paths and the old database remains usable.
- **AC-012**: Given rollback cannot restore consistent active roots, when startup continues, then Miru does not open SQLite and shows a blocking initialization error.
- **AC-013**: Given fresh initialization fails, when Miru reports the failure, then reset backups and the marker remain available for a later retry.
- **AC-014**: Given backup cleanup fails, when the fresh app opens, then the About panel reports incomplete cleanup and its size includes remaining backups.
- **AC-015**: Given cleanup was incomplete, when a later startup can remove every backup, then the marker is removed and the warning clears.
- **AC-016**: Given reset completes, when the user views data size, then the displayed value represents only files created by the fresh startup and is not required to be literal `0 B`.

## 6. Test Automation Strategy

- **Test levels**: Unit tests for path validation, sizing, marker transitions, staging, rollback, cleanup, and App guards; frontend type checking and linting; manual Wails UI verification for focus and real quit behavior.
- **Frameworks**: Go `testing`, temporary directories, existing React/TypeScript tooling, and existing project Make targets.
- **Test data management**: Every filesystem test shall use `t.TempDir()` and synthetic config, cache, data, marker, backup, symlink, and permission-error fixtures. Tests shall never operate on resolved user directories.
- **Failure injection**: Filesystem operations shall be structured so tests can exercise rename and cleanup failures without adding a general-purpose abstraction layer. Platform-specific permission tests may be skipped when the operating system cannot express the condition reliably.
- **Crash recovery**: Unit tests shall begin from each valid marker phase and from partially completed root states to verify idempotent recovery.
- **CI/CD integration**: Run `make fmt`, `make lint`, `make test`, `make fmt-fe`, `make typecheck`, and `make lint-fe`. Run `make check-windows` when practical because open-file behavior is a primary design constraint.
- **Coverage requirements**: Valid marker phases, successful rollback, malformed marker handling, cleanup retry, duplicate roots, and symlink safety shall have direct tests. Platform-specific failure branches may rely on static checks when the operating system cannot express them reliably. No numeric repository-wide coverage threshold is introduced.
- **Performance testing**: Size scanning shall be tested with nested directories. The Settings view shall show a loading state while scanning; no fixed file-count benchmark is required for this local filesystem operation.
- **Manual verification gaps**: Verify the running Wails app at `http://127.0.0.1:34115`; confirm dialog focus, disabled playback state, app quit, manual reopen, fresh UI, retained downloaded media, and cleanup warning presentation.

## 7. Rationale & Context

Miru cannot reliably delete its live database and log files on every supported operating system. SQLite may own WAL and shared-memory files, and the Wails logger keeps the active log open. Running the reset during the next process startup removes those handles from the old roots.

Renaming a directory within its parent is the smallest useful staging operation and preserves all files, including SQLite sidecars, without enumerating them. Each Miru root may be on a different filesystem, so the marker records progress and enables rollback across the three independent renames.

The application initializes a new SQLite database and log immediately after reset. A literal `0 B` display after reopening would be inaccurate. The UI must show the current fresh-state size instead.

The design reuses the existing dialog and destructive button conventions. A new confirmation framework or dependency is unnecessary. Cancel receives safe initial focus because Enter must not trigger permanent deletion by accident.

## 8. Dependencies & External Integrations

### External Systems

- **EXT-001**: Operating-system filesystem APIs for metadata, rename, directory creation, and recursive deletion.
- **EXT-002**: Wails runtime quit lifecycle for a normal application shutdown.

### Third-Party Services

- **SVC-001**: None. Reset must not require network access.

### Infrastructure Dependencies

- **INF-001**: Writable parent directories for the resolved config, cache, and data roots.
- **INF-002**: Sufficient free directory entries and permissions to create marker, backup names, and fresh roots.

### Data Dependencies

- **DAT-001**: `paths.Dirs` is the sole authority for active Miru filesystem roots.
- **DAT-002**: Known frontend persistence keys are `miru.playback` and `miru.search`.

### Technology Platform Dependencies

- **PLT-001**: Wails v2 application lifecycle and runtime bindings.
- **PLT-002**: Go standard-library filesystem and JSON support.
- **PLT-003**: React, Zustand persistence, Tailwind tokens, and existing shadcn-style primitives in `frontend/`.

### Compliance Dependencies

- **COM-001**: No formal compliance regime is claimed. The UI shall not describe the action as secure erasure.

## 9. Examples & Edge Cases

### Successful lifecycle

```text
idle -> marker pending -> normal quit
next launch -> config/cache/data staged -> fresh roots created
fresh SQLite initialized -> backups removed -> marker removed -> fresh state
```

### Recoverable staging failure

```text
config staged -> cache rename fails -> config rollback succeeds
result: old active roots restored, old database usable, reset error shown
```

### Cleanup failure

```text
all roots staged -> fresh SQLite initialized -> cache backup deletion fails
result: fresh active state remains, marker becomes cleanup_pending,
        remaining backup size is shown, cleanup retries on a later startup
```

### Required edge cases

- One or more active roots do not exist.
- An active root is empty.
- A backup path already exists for the Reset ID.
- The marker contains malformed JSON, an unsupported schema version, or an unknown root name.
- The process exits after a root rename but before the marker update.
- Both an active path and its expected backup path exist.
- Config and data resolve to the same directory.
- A root contains symbolic links to user files outside Miru directories.
- The database uses `app_data.db-wal` and `app_data.db-shm`.
- The logger has created a new file before cleanup is requested from `App.startup`.
- Data size exceeds 2 GB.
- `localStorage` contains unrelated keys. Reset removes only known Miru keys.
- MPV exits between frontend disabled-state evaluation and the backend check.
- A torrent is downloading, paused, or seeding when reset is confirmed.

## 10. Validation Criteria

The implementation complies with this specification only when:

- Every acceptance criterion has an automated test or an explicitly documented manual-only reason.
- All required Make checks pass without editing generated Wails files manually.
- Path-safety tests prove no reset operation can target a parent, home, download, or unrelated directory.
- Crash-recovery tests prove marker processing is idempotent for every valid phase and partial staging state.
- A staging failure leaves the prior SQLite database usable when rollback succeeds.
- A cleanup failure is visible and remaining backup bytes are included in the displayed size.
- Keyboard-only testing confirms safe focus, Escape cancellation, visible focus, and disabled-state behavior.
- Manual Wails verification confirms that downloaded media survives a complete reset.
- README documentation describes the reset scope, restart requirement, included data, and download exclusion.

## 11. Related Specifications / Further Reading

- [GitHub issue #44](https://github.com/sunnyegg/miru/issues/44)
- [`AGENTS.md`](../AGENTS.md)
- [`internal/paths/paths.go`](../internal/paths/paths.go)
- [`internal/storage/storage.go`](../internal/storage/storage.go)
- [`frontend/src/views/Settings.tsx`](../frontend/src/views/Settings.tsx)
- [`frontend/src/components/settings/SettingsAboutPanel.tsx`](../frontend/src/components/settings/SettingsAboutPanel.tsx)
