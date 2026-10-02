import {useMemo, useState} from 'react'
import type {ShowGroup} from '../lib/groupEpisodes'
import {
  bucketOnDiskShows,
  onDiskFilters,
  type OnDiskFilter,
} from '../lib/libraryOnDisk'
import type {WatchingEntryView} from '../lib/types'
import {LibraryPosterGrid} from './LibraryPosterGrid'
import {Alert, AlertDescription} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'

type Props = {
  loading: boolean
  loadError: string
  statusError: string
  shows: ShowGroup[]
  entries: WatchingEntryView[]
  highlightedKey: string | null
  onSelectShow: (key: string) => void
  onRetry: () => void
  onRetryLibrary: () => void
  suppressEmptyState?: boolean
}

export function LibraryOnDiskSection({
  loading,
  loadError,
  statusError,
  shows,
  entries,
  highlightedKey,
  onSelectShow,
  onRetry,
  onRetryLibrary,
  suppressEmptyState = false,
}: Props) {
  const [selectedFilter, setSelectedFilter] = useState<OnDiskFilter>('ALL')
  const buckets = useMemo(
    () => bucketOnDiskShows(shows, entries),
    [entries, shows],
  )

  const availableFilters = statusError ? [onDiskFilters[0]] : onDiskFilters
  const activeFilter = statusError ? 'ALL' : selectedFilter

  if (!loading && shows.length === 0 && suppressEmptyState && !statusError) {
    return null
  }

  return (
    <section className="shrink-0">
      <div className="mb-3 flex items-baseline gap-2">
        <h3 className="text-sm font-medium text-foreground">On disk</h3>
        {!loading && (
          <span className="text-xs text-muted-foreground">{shows.length}</span>
        )}
      </div>

      {statusError && (
        <Alert
          variant="destructive"
          className="mb-4 flex flex-wrap items-center justify-between gap-4 p-4"
        >
          <div className="min-w-0">
            <p className="font-medium text-foreground">
              AniList statuses could not be loaded
            </p>
            <AlertDescription className="mt-1">
              All local shows remain visible. {statusError}
            </AlertDescription>
          </div>
          <Button type="button" variant="secondary" onClick={onRetry}>
            Try again
          </Button>
        </Alert>
      )}

      {!loading && shows.length > 0 && (
        <div
          className="mb-4 flex flex-wrap gap-2"
          role="group"
          aria-label="Filter on-disk shows"
        >
          {availableFilters.map(({value, label}) => {
            const active = value === activeFilter
            const count = buckets[value].length
            return (
              <Button
                key={value}
                type="button"
                variant="ghost"
                aria-pressed={active}
                onClick={() => setSelectedFilter(value)}
                className={
                  active
                    ? 'border-accent bg-muted text-foreground'
                    : 'text-muted-foreground'
                }
              >
                {label}
                <span className="text-xs tabular-nums text-muted-foreground">
                  {count}
                </span>
              </Button>
            )
          })}
        </div>
      )}

      <LibraryPosterGrid
        loading={loading}
        loadError={loadError}
        shows={buckets[activeFilter]}
        highlightedKey={highlightedKey}
        onSelectShow={onSelectShow}
        onRetry={onRetryLibrary}
        suppressEmptyState={suppressEmptyState && activeFilter === 'ALL'}
        emptyMessage={
          activeFilter === 'ALL'
            ? undefined
            : `No on-disk shows with ${onDiskFilters.find((filter) => filter.value === activeFilter)?.label ?? 'this status'}.`
        }
      />
    </section>
  )
}
