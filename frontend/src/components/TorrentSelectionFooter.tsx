import {useEffect, useRef} from 'react'
import {Button} from '@/components/ui/button'

type Progress = {
  inspecting: boolean
  position: number
  total: number
}

type Props = {
  selectMode: boolean
  selectedCount: number
  allSelected: boolean
  someSelected: boolean
  sizeLabel: string | null
  running: boolean
  progress: Progress | null
  onToggleAll: () => void
  onLeaveMode: () => void
  onStartBatch: () => void
  onCancelBatch: () => void
}

// TorrentSelectionFooter sits below the result list, so selection mode adds no
// chrome above it: the list stays the hero and the decision waits at the bottom
// where the eye already lands after scrolling. During a batch the same bar
// becomes the progress line.
export function TorrentSelectionFooter({
  selectMode,
  selectedCount,
  allSelected,
  someSelected,
  sizeLabel,
  running,
  progress,
  onToggleAll,
  onLeaveMode,
  onStartBatch,
  onCancelBatch,
}: Props) {
  const allRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (allRef.current) {
      allRef.current.indeterminate = someSelected && !allSelected
    }
  }, [someSelected, allSelected])

  if (running && progress) {
    return (
      <div className="flex min-h-15 shrink-0 items-center gap-3 border border-border bg-card px-3">
        <p className="flex min-h-11 min-w-0 flex-1 items-center text-sm text-muted-foreground">
          <span aria-live="polite" className="truncate tabular-nums">
            {progress.inspecting
              ? `Inspecting ${progress.position} of ${progress.total}…`
              : `Torrent ${progress.position} of ${progress.total}`}
          </span>
        </p>
        <Button type="button" variant="ghost" onClick={onCancelBatch}>
          Cancel batch
        </Button>
      </div>
    )
  }

  if (!selectMode) {
    return null
  }

  return (
    <div className="flex min-h-15 shrink-0 flex-wrap items-center gap-3 border border-border bg-card px-3">
      <p className="min-w-0 flex-1 text-sm">
        <span className="text-base font-semibold">
          {selectedCount} {selectedCount === 1 ? 'torrent' : 'torrents'}
        </span>
        <span className="ml-2 text-muted-foreground">
          {selectedCount > 0 ? (sizeLabel ?? 'Size unknown') : ''}
        </span>
      </p>
      <label className="flex min-h-11 items-center gap-2 text-sm">
        <input
          ref={allRef}
          type="checkbox"
          className="size-4 accent-primary"
          checked={allSelected}
          onChange={onToggleAll}
        />
        Select all
      </label>
      <Button type="button" variant="ghost" onClick={onLeaveMode}>
        Cancel
      </Button>
      <Button
        type="button"
        onClick={onStartBatch}
        disabled={selectedCount === 0}
      >
        Download {selectedCount}
      </Button>
    </div>
  )
}
