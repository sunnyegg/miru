# Spec: multi-select torrent download

Batch-download several torrents from the indexer search results in one action, instead of
clicking Download and confirming a file sheet once per result.

Status: agreed, implemented.

## Goals

- Select multiple results on the current page in the Torrent → Indexer search list (Nyaa and
  Tokyo Toshokan share the same list) and add them all to the download queue.
- Keep the existing per-torrent file picker (`TorrentFileSheet`), shown once per selected
  torrent in card order.
- Zero backend change: batch relies on the queue that already exists (`Manager.Start` inserts a
  `QUEUED` job, `PumpQueue` respects `maxConcurrentDownloads`).

## Non-goals

- No multi-select in RSS feeds (`FeedSubscriptions`), Downloads history, or the Library
  "Find torrent" flow.
- No batch concept in the data model: every torrent is still one row in Downloads.
- No duplicate detection. `internal/storage` has no lookup by `source`; re-adding a torrent that
  is already queued keeps working exactly like the single Download button does today.
- No new Wails binding, no `internal/torrentx` change, no `wailsjs` regeneration.

## Decisions

Every item below was settled during the design grilling and revised once against the running app;
keep them in step with the code.

1. Scope is the indexer search result list only.
2. Selection lives in local state in `SearchView`, keyed by `magnet || link` (not by index). It
   survives page changes (so `Select all` can mean every result) and resets when a new
   search/source replaces the results or when selection mode is left.
3. Selection mode is explicit: a `Select` button beside the search field turns it on, and
   `Cancel` in the footer turns it off, clearing the selection. Checkboxes only exist in that
   mode, so the default list looks exactly like it did before.
4. No chrome is added above the list. The count, the total, `Select all`, `Cancel` and
   `Download N` live in a footer below the results, so the list stays the hero and the decision
   sits where the eye lands after scrolling.
5. Batch skips no torrent: the file sheet is shown per torrent, in card order, so the existing
   per-file choice is preserved. File selection defaults to all files.
6. `InspectTorrent` is prefetched for all selected torrents when the batch starts, with a
   concurrency cap of 4, in card order (`PREFETCH_CONCURRENCY = 4`).
7. The first sheet opens as soon as its own inspect finishes; remaining prefetches continue in
   the background. A sheet whose data is not ready yet opens in its loading state with the
   confirm button disabled.
8. An inspect failure does not remove the torrent from the batch. Its sheet still opens, showing
   the error inline with `Skip` and `Retry`, so failures are handled by one rule anywhere in the
   flow. `Skip` and `Retry` appear only in the error state, never while loading.
9. Closing the sheet (X, Escape, backdrop) cancels the rest of the batch. No confirmation and no
   rollback; torrents already added stay in the queue.
10. Selection, checkboxes and the mode toggle are locked while a batch runs.
11. During a batch the footer shows progress (`Inspecting 2 of 5…`, `Torrent 2 of 5`) and a
    `Cancel batch` button. If every prefetch fails, no sheet opens at all: stay on Search, error
    toast, selection untouched.
12. `Select all` is one checkbox with that label, sitting in the footer, and it covers every
    result, not just the current page; it shows the indeterminate state when only some results
    are selected. There is no `Clear` action: leaving the mode clears the selection.
13. `parseSize` is added to `lib/format.ts`: accepts `B`, `KiB`, `MiB`, `GiB`, `TiB` and `KB`,
    `MB`, `GB`, `TB` (all treated as 1024-based so the total cannot surprise), decimal `.` or
    `,`, optional space, case-insensitive; anything else returns `null`.
14. The footer shows `N torrents` with the summed size (`3 torrents · 2.4 GB`) formatted by the
    existing `formatBytes` (decimal labels, 1024 base, same as Downloads and the file sheet).
    When some sizes cannot be parsed the total is prefixed `~`; when none can, only the count is
    shown. Which item failed to parse is not marked.
15. Clicking a card toggles its selection, but only in selection mode.
16. Card click guards: no toggle when a text selection is active, when the click target is inside
    `button`, `input`, `label` or `a`, or while a batch is running.
17. A selected card gets an `accent` border with a 2px `accent` left edge, and `cursor-pointer`
    in selection mode; `cursor-default` otherwise. No new hex colours or radii.
18. Checkbox `aria-label` is `Select <torrent title>` with the title truncated to ~80 characters.
19. Switching tabs mid-batch unmounts `SearchView` and ends the batch: prefetches that are
    already running finish and no-op, sheets never open, added jobs stay. Batch code must not
    call `setState` after unmount.
20. The per-card `Download` button is hidden in selection mode, because the footer's
    `Download N` is the only way to start a download there.
21. Docs change in the same change: `README.md` Search bullet and `AGENTS.md` frontend notes.

## UX flow

1. `Select` beside the search field → mode on: every card gains a checkbox at its right edge,
   its `Download` button disappears, and the footer appears with `0 torrents`, the `Select all`
   checkbox, `Cancel` and a disabled `Download 0`.
2. Check cards (or click them), or tick `Select all` → `N torrents · 2.4 GB` (`~` when some sizes
   are unreadable) and `Download N` enabled.
3. `Download N` → selection locks, prefetch starts (cap 4, card order), the footer switches to
   progress.
4. Sheet 1 opens when ready (`Torrent 1 of 5` + the torrent name), file list with all files
   selected, confirm button `Download`.
5. Confirm → `StartTorrent(source, files)` → next sheet opens automatically. Error → `Skip` /
   `Retry`.
6. End of batch → toast, jump to Downloads, selection cleared.
7. `Cancel batch` or closing a sheet → stop the rest, stay on Search.
8. `Cancel` in the footer leaves the mode and clears the selection.

## Messages

| Case | Message | Navigation |
| --- | --- | --- |
| All added | `5 downloads added` | Downloads |
| Some failed | `4 downloads added, 1 failed` (warning, not error) | Downloads |
| All failed | error, e.g. `Could not read torrent files` | stay on Search, selection kept |
| Cancelled mid-batch | `2 downloads added` | stay on Search |
| Single result without link or magnet | counts as a prefetch failure | — |

Items that succeed after `Retry` count as added.

## Files

| File | Change |
| --- | --- |
| `frontend/src/lib/format.ts` | add `parseSize` (~15 lines, no dependency) |
| `frontend/src/lib/torrentBatch.ts` | new: `PREFETCH_CONCURRENCY`, ordered prefetch pool, the batch item/state transitions (`batchItemsFor`, `inspectSettled`, `markCurrent`, `toggleKey(s)`), `summarizeBatch`, total-size rule; pure, no React import |
| `frontend/src/views/useTorrentBatch.ts` | new: batch state machine — selection keys, prefetch start and settle, `resolve`/`retry`/`skip`/`cancel`, token and unmount guards, the four end-of-batch cases; co-located with its only consumer, not a `hooks/` folder |
| `frontend/src/components/TorrentSelectionFooter.tsx` | new: the bar below the result list — `N torrents` with the total, the `Select all` checkbox (indeterminate when partial), `Cancel` to leave the mode, `Download N`, and the `aria-live="polite"` progress line with `Cancel batch` during a batch |
| `frontend/src/components/TorrentResultCard.tsx` | checkbox (44px hit area) at the card's right edge that only exists in selection mode, `accent` border plus a 2px left edge when selected, card-click toggle with guards, `Download` button hidden in selection mode |
| `frontend/src/components/TorrentFileSheet.tsx` | position label, `Skip`/`Retry` in the error state, confirm disabled while loading, "closing stops the rest" note |
| `frontend/src/views/Search.tsx` | renders the bar, cards and sheet from the hook, plus the page slice, selection totals and pagination |

`Search.tsx` was 384 lines before this change. The batch state machine lives in
`frontend/src/views/useTorrentBatch.ts`, co-located with its only consumer: the state stays owned by
the view (local `useState`, not a store), while the async orchestration (Wails calls, notices,
navigation) and the four end-of-batch cases sit in the hook, pure logic in `lib/torrentBatch.ts`,
and the mode footer in its own component. No `hooks/` folder (nothing is reused across screens).

## Verification

- `make fmt-fe`, `make typecheck`, `make lint-fe`. No Go change, so `make test` / `make lint` are
  not required.
- No frontend test runner exists (no vitest in `frontend/package.json`), so `parseSize` and the
  batch helpers are verified by typecheck plus the manual pass below.
- Manual/visual with `make dev` at `http://127.0.0.1:34115` (mode `interaction`): select 3 results
  → bar appears with total → download → sheets in order → toast → Downloads. Also: page change
  resets selection, dead magnet shows the error sheet and `Skip` continues, closing a sheet stops
  the rest, tab switch mid-batch ends the batch quietly.
- Design-system check: icons from `components/Icons.tsx` only, no new hex colours or radii,
  44px controls, purple only as the hit (`Download N`, `accent` selection border).
