import {errorMessage, formatBytes, parseSize} from './format'
import type {NyaaResultView, TorrentContentsView} from './types'

export const PREFETCH_CONCURRENCY = 4

// A selected result without a link or magnet still joins the batch so one
// Skip/Retry rule in the sheet covers every torrent.
export const TORRENT_SOURCE_MISSING = 'This result has no torrent link'

export type TorrentBatchItem = {
  key: string
  title: string
  source: string | null
  status: 'pending' | 'ready' | 'failed' | 'added'
  contents: TorrentContentsView | null
  error: string
}

export type TorrentBatchState = {
  items: TorrentBatchItem[]
  index: number
  sheetOpen: boolean
  // bulk is false for a single card Download: no position label, and it ends
  // with the single-torrent message instead of the batch summary.
  bulk: boolean
}

export type PrefetchResult =
  | {status: 'ready'; contents: TorrentContentsView | null}
  | {status: 'failed'; error: string}

// A selection is keyed by magnet || link, never by index, so it survives the
// list re-rendering but not a new page or a new search.
export function torrentKey(result: NyaaResultView): string {
  return result.magnet || result.link
}

export function batchItemsFor(
  results: readonly NyaaResultView[],
): TorrentBatchItem[] {
  return results.map((result) => {
    const source = result.link || result.magnet || null
    return {
      key: torrentKey(result),
      title: result.title,
      source,
      status: source ? 'pending' : 'failed',
      contents: null,
      error: source ? '' : TORRENT_SOURCE_MISSING,
    }
  })
}

export function toggleKey(selected: readonly string[], key: string): string[] {
  return selected.includes(key)
    ? selected.filter((entry) => entry !== key)
    : [...selected, key]
}

// toggleKeys selects a whole page at once, or clears it when all of it is
// already selected.
export function toggleKeys(
  selected: readonly string[],
  keys: readonly string[],
): string[] {
  const missing = keys.filter((key) => !selected.includes(key))
  if (missing.length > 0) {
    return [...selected, ...missing]
  }
  return selected.filter((key) => !keys.includes(key))
}

export function withItemPatch(
  items: readonly TorrentBatchItem[],
  index: number,
  patch: Partial<TorrentBatchItem>,
): TorrentBatchItem[] {
  return items.map((item, itemIndex) =>
    itemIndex === index ? {...item, ...patch} : item,
  )
}

// inspectSettled stores one inspect result. The sheet opens once the torrent it
// would show has settled, so a prefetch that finishes early never reorders the
// sheets.
export function inspectSettled(
  state: TorrentBatchState,
  index: number,
  result: PrefetchResult,
): TorrentBatchState {
  const item = state.items[index]
  if (!item) {
    return state
  }
  // A result with no source is already failed when the batch is built, and a
  // settled index that is no longer current must keep its outcome. Either way
  // the sheet still opens, otherwise a source-less first item would leave the
  // batch waiting for a sheet that never comes.
  const items =
    item.status === 'pending'
      ? withItemPatch(state.items, index, {
          status: result.status === 'ready' ? 'ready' : 'failed',
          contents: result.status === 'ready' ? result.contents : null,
          error: result.status === 'ready' ? '' : result.error,
        })
      : state.items
  return {
    ...state,
    items,
    sheetOpen: state.sheetOpen || state.index === index,
  }
}

// markCurrent records the open sheet's outcome and moves to the next torrent.
// done is true when that was the last one and the batch is over.
export function markCurrent(
  state: TorrentBatchState,
  patch: Partial<TorrentBatchItem>,
): {state: TorrentBatchState; done: boolean} {
  const items = withItemPatch(state.items, state.index, patch)
  if (state.index + 1 >= items.length) {
    return {state: {...state, items}, done: true}
  }
  return {
    state: {...state, items, index: state.index + 1, sheetOpen: true},
    done: false,
  }
}

export function addedLabel(count: number): string {
  return `${count} download${count === 1 ? '' : 's'} added`
}

// prefetchTorrents inspects the sources in card order with at most
// PREFETCH_CONCURRENCY in flight. Each result is reported as soon as it settles.
// Once isStopped reports true no new inspect starts and late results are dropped.
export async function prefetchTorrents(
  sources: readonly (string | null)[],
  inspect: (source: string) => Promise<TorrentContentsView>,
  onSettled: (index: number, result: PrefetchResult) => void,
  isStopped: () => boolean,
): Promise<void> {
  let nextIndex = 0

  async function worker() {
    while (nextIndex < sources.length) {
      if (isStopped()) {
        return
      }
      const index = nextIndex
      nextIndex += 1
      const source = sources[index]
      if (!source) {
        onSettled(index, {status: 'failed', error: TORRENT_SOURCE_MISSING})
        continue
      }
      let result: PrefetchResult
      try {
        result = {status: 'ready', contents: await inspect(source)}
      } catch (err) {
        result = {status: 'failed', error: errorMessage(err)}
      }
      if (isStopped()) {
        return
      }
      onSettled(index, result)
    }
  }

  const workers = Math.min(PREFETCH_CONCURRENCY, sources.length)
  await Promise.all(Array.from({length: workers}, () => worker()))
}

export function summarizeBatch(items: readonly TorrentBatchItem[]): {
  added: number
  failed: number
} {
  let added = 0
  let failed = 0
  for (const item of items) {
    if (item.status === 'added') {
      added += 1
    } else if (item.status === 'failed') {
      failed += 1
    }
  }
  return {added, failed}
}

// batchSizeLabel sums the sizes it can read and returns null when none can, so
// the bar can fall back to the bare count. A partial sum is prefixed with ~.
export function batchSizeLabel(sizes: readonly string[]): string | null {
  const parsed = sizes.map(parseSize)
  const known = parsed.filter((size): size is number => size !== null)
  if (known.length === 0) {
    return null
  }
  const total = known.reduce((sum, size) => sum + size, 0)
  const label = formatBytes(total)
  return known.length === parsed.length ? label : `~${label}`
}
