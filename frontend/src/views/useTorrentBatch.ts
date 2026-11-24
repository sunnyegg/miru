import {useEffect, useRef, useState} from 'react'
import {InspectTorrent, StartTorrent} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import {
  TORRENT_SOURCE_MISSING,
  addedLabel,
  batchItemsFor,
  inspectSettled,
  markCurrent,
  prefetchTorrents,
  summarizeBatch,
  toggleKey,
  toggleKeys,
  torrentKey,
  withItemPatch,
  type PrefetchResult,
  type TorrentBatchItem,
  type TorrentBatchState,
} from '../lib/torrentBatch'
import type {NyaaResultView, TorrentFileView} from '../lib/types'

type NoticeFn = (message: string, isError?: boolean) => void

type Options = {
  notice: NoticeFn
  goToDownloads: () => void
  results: readonly NyaaResultView[]
}

export type TorrentBatchController = {
  selectMode: boolean
  selected: string[]
  batch: TorrentBatchState | null
  shownItem: TorrentBatchItem | undefined
  startingKey: string | undefined
  confirming: boolean
  toggleMode: () => void
  toggleSelected: (result: NyaaResultView) => void
  toggleAll: (allResults: readonly NyaaResultView[]) => void
  downloadSelected: (allResults: readonly NyaaResultView[]) => void
  downloadOne: (result: NyaaResultView) => void
  confirmBatch: (files: TorrentFileView[]) => void
  skipCurrent: () => void
  retryCurrent: () => Promise<void>
  cancelBatch: () => void
}

// useTorrentBatch owns indexer-result selection and the batch that adds the
// selected torrents one file sheet at a time. Pure transitions live in
// lib/torrentBatch.ts; this hook adds the Wails calls, the notices and the
// navigation the four end-of-batch cases need.
export function useTorrentBatch({
  notice,
  goToDownloads,
  results,
}: Options): TorrentBatchController {
  const [selectMode, setSelectMode] = useState(false)
  const [selected, setSelected] = useState<string[]>([])
  const [batch, setBatch] = useState<TorrentBatchState | null>(null)
  const [confirming, setConfirming] = useState(false)
  // batchRef mirrors batch so prefetch callbacks and sheet actions read the
  // latest batch, and so finishBatch can run outside a state updater.
  const batchRef = useRef<TorrentBatchState | null>(null)
  const batchToken = useRef<object | null>(null)
  const mounted = useRef(true)

  // A tab switch unmounts the view: running prefetches finish but their results
  // are dropped and no sheet opens.
  useEffect(() => {
    mounted.current = true
    return () => {
      mounted.current = false
      batchToken.current = null
    }
  }, [])

  useEffect(() => {
    setSelected([])
  }, [results])

  function commitBatch(next: TorrentBatchState | null) {
    batchRef.current = next
    if (mounted.current) {
      setBatch(next)
    }
  }

  function stopped(token: object) {
    return batchToken.current !== token || !mounted.current
  }

  function startBatch(items: TorrentBatchItem[], bulk: boolean) {
    if (items.length === 0 || batchRef.current) {
      return
    }
    const token = {}
    batchToken.current = token
    commitBatch({items, index: 0, sheetOpen: false, bulk})
    void prefetchTorrents(
      items.map((item) => item.source),
      InspectTorrent,
      (index, result) => settleInspect(token, index, result),
      () => stopped(token),
    )
  }

  function settleInspect(token: object, index: number, result: PrefetchResult) {
    const current = batchRef.current
    if (stopped(token) || !current) {
      return
    }
    commitBatch(inspectSettled(current, index, result))
  }

  // resolveCurrent records the open sheet's outcome and moves to the next
  // torrent, or ends the batch when it was the last one.
  function resolveCurrent(token: object, patch: Partial<TorrentBatchItem>) {
    const current = batchRef.current
    if (stopped(token) || !current) {
      return
    }
    const {state, done} = markCurrent(current, patch)
    if (!done) {
      commitBatch(state)
      return
    }
    finishBatch(state)
  }

  function finishBatch(finished: TorrentBatchState) {
    const {added, failed} = summarizeBatch(finished.items)
    batchToken.current = null
    commitBatch(null)
    setConfirming(false)
    if (!finished.bulk) {
      if (added > 0) {
        notice('Download added')
        goToDownloads()
      }
      return
    }
    if (failed === 0) {
      setSelected([])
      notice(addedLabel(added))
      goToDownloads()
    } else if (added === 0) {
      // Every torrent failed: keep the selection so a retry starts here.
      notice('Could not read torrent files', true)
    } else {
      setSelected([])
      notice(`${addedLabel(added)}, ${failed} failed`)
      goToDownloads()
    }
  }

  // Closing the sheet, Cancel batch or leaving the tab drops the remaining
  // torrents; the ones already added stay in the queue.
  function cancelBatch() {
    const current = batchRef.current
    batchToken.current = null
    commitBatch(null)
    setConfirming(false)
    if (!current) {
      return
    }
    const {added} = summarizeBatch(current.items)
    if (added > 0) {
      notice(addedLabel(added))
    }
  }

  function skipCurrent() {
    const token = batchToken.current
    if (token) {
      resolveCurrent(token, {status: 'failed', error: 'Skipped'})
    }
  }

  async function retryCurrent() {
    const token = batchToken.current
    const current = batchRef.current
    if (!token || !current) {
      return
    }
    const index = current.index
    const torrentSource = current.items[index].source
    if (!torrentSource) {
      return
    }
    const pending: Partial<TorrentBatchItem> = {status: 'pending', error: ''}
    commitBatch({
      ...current,
      items: withItemPatch(current.items, index, pending),
    })
    try {
      const contents = await InspectTorrent(torrentSource)
      settleInspect(token, index, {status: 'ready', contents})
    } catch (err) {
      settleInspect(token, index, {status: 'failed', error: errorMessage(err)})
    }
  }

  async function confirmBatch(files: TorrentFileView[]) {
    const token = batchToken.current
    const current = batchRef.current
    if (!token || !current) {
      return
    }
    const torrentSource = current.items[current.index].source
    if (!torrentSource) {
      return
    }
    setConfirming(true)
    try {
      await StartTorrent(torrentSource, files)
    } catch (err) {
      if (mounted.current) {
        notice(errorMessage(err), true)
        setConfirming(false)
      }
      return
    }
    if (!mounted.current) {
      return
    }
    setConfirming(false)
    resolveCurrent(token, {status: 'added', error: ''})
  }

  function downloadOne(result: NyaaResultView) {
    if (!result.link && !result.magnet) {
      notice(TORRENT_SOURCE_MISSING, true)
      return
    }
    startBatch(batchItemsFor([result]), false)
  }

  function downloadSelected(allResults: readonly NyaaResultView[]) {
    const chosen = allResults.filter((result) =>
      selected.includes(torrentKey(result)),
    )
    startBatch(batchItemsFor(chosen), true)
  }

  // Leaving selection mode drops the selection, so re-entering always starts
  // from a clean slate instead of resurrecting keys from an older query.
  function toggleMode() {
    setSelectMode((current) => !current)
    setSelected([])
  }

  function toggleSelected(result: NyaaResultView) {
    setSelected((current) => toggleKey(current, torrentKey(result)))
  }

  function toggleAll(allResults: readonly NyaaResultView[]) {
    setSelected((current) =>
      toggleKeys(
        current,
        allResults.map((result) => torrentKey(result)),
      ),
    )
  }

  const shownItem = batch?.sheetOpen ? batch.items[batch.index] : undefined
  const startingKey =
    batch && !batch.bulk && batch.items[0]?.status === 'pending'
      ? batch.items[0].key
      : undefined

  return {
    selectMode,
    selected,
    batch,
    shownItem,
    startingKey,
    confirming,
    toggleMode,
    toggleSelected,
    toggleAll,
    downloadSelected,
    downloadOne,
    confirmBatch,
    skipCurrent,
    retryCurrent,
    cancelBatch,
  }
}
