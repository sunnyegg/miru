import {useEffect, useState} from 'react'
import {IconSearch} from '../components/Icons'
import {WatchingEditSheet} from '../components/WatchingEditSheet'
import {WatchingSearchDialog} from '../components/WatchingSearchDialog'
import type {AnimeListEntryInput} from '../lib/types'
import {useNavigationStore} from '../stores/navigationStore'
import {useWatchingStore, type ListFilter} from '../stores/watchingStore'
import {Alert, AlertAction, AlertDescription} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'
import {Progress} from '@/components/ui/progress'
import {Skeleton} from '@/components/ui/skeleton'
import {cn} from '@/lib/utils'

type Props = {
  notice: (msg: string, isError?: boolean) => void
}

const listFilters: {value: ListFilter; label: string}[] = [
  {value: 'CURRENT', label: 'Watching'},
  {value: 'COMPLETED', label: 'Completed'},
  {value: 'PLANNING', label: 'Planning'},
  {value: 'PAUSED', label: 'Paused'},
  {value: 'DROPPED', label: 'Dropped'},
  {value: 'REPEATING', label: 'Repeating'},
]

function listStatusLabel(status: string): string {
  const match = listFilters.find((filter) => filter.value === status)
  if (match) {
    return match.label
  }
  return status
}

const mediaStatusLabels: Record<string, string> = {
  RELEASING: 'Airing',
  FINISHED: 'Finished',
  NOT_YET_RELEASED: 'Not yet aired',
  CANCELLED: 'Cancelled',
  HIATUS: 'On hiatus',
}

function mediaStatusLabel(status: string): string {
  if (!status) {
    return ''
  }
  return mediaStatusLabels[status] ?? status
}

export function WatchingView({notice}: Props) {
  const goToSettings = useNavigationStore((state) => state.setTab)
  const listFilter = useWatchingStore((state) => state.listFilter)
  const entries = useWatchingStore((state) => state.entries)
  const counts = useWatchingStore((state) => state.counts)
  const countsLoading = useWatchingStore((state) => state.countsLoading)
  const countsError = useWatchingStore((state) => state.countsError)
  const loading = useWatchingStore((state) => state.loading)
  const notConnected = useWatchingStore((state) => state.notConnected)
  const error = useWatchingStore((state) => state.error)
  const selectFilter = useWatchingStore((state) => state.selectFilter)
  const loadList = useWatchingStore((state) => state.loadList)
  const loadCounts = useWatchingStore((state) => state.loadCounts)
  const saveEntry = useWatchingStore((state) => state.saveEntry)

  const [searchOpen, setSearchOpen] = useState(false)
  const [editingEntry, setEditingEntry] = useState<
    (typeof entries)[number] | null
  >(null)
  const [savingEntry, setSavingEntry] = useState(false)

  useEffect(() => {
    void loadList()
    void loadCounts()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function saveEntryWithState(input: AnimeListEntryInput) {
    setSavingEntry(true)
    try {
      await saveEntry(input, notice)
      setEditingEntry(null)
    } catch {
      // notice handled in store
    } finally {
      setSavingEntry(false)
    }
  }

  const filterLabel = listStatusLabel(listFilter)
  const emptyCopy = `Nothing on your ${filterLabel} list. Switch lists, or search AniList to add a title.`

  return (
    <section className="flex h-full flex-col gap-6">
      <header>
        <p className="text-xs font-medium tracking-[0.16em] text-muted-foreground uppercase">
          AniList
        </p>
        <div className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <h2 className="text-3xl font-semibold tracking-tight">
              Anime list
            </h2>
            <p className="mt-1 text-sm text-muted-foreground">
              Track what you are watching and where you left off.
            </p>
          </div>
          {!notConnected && (
            <Button
              type="button"
              variant="secondary"
              onClick={() => setSearchOpen(true)}
            >
              <IconSearch className="size-4" />
              Find a title
            </Button>
          )}
        </div>

        <div
          className="mt-5 -mx-1 flex gap-1 overflow-x-auto border-b border-border px-1"
          role="group"
          aria-label="Anime list status"
        >
          {listFilters.map((filter) => {
            const selected = listFilter === filter.value
            return (
              <Button
                key={filter.value}
                type="button"
                variant="ghost"
                className={cn(
                  'relative h-11 min-h-11 shrink-0 px-3 text-muted-foreground hover:text-foreground',
                  selected && 'text-foreground',
                )}
                aria-pressed={selected}
                onClick={() => void selectFilter(filter.value)}
              >
                {filter.label}
                {!countsLoading && !countsError && (
                  <span className="text-xs tabular-nums text-muted-foreground">
                    {counts[filter.value] ?? 0}
                  </span>
                )}
                {selected && (
                  <span
                    className="absolute inset-x-3 bottom-0 h-0.5 bg-accent"
                    aria-hidden="true"
                  />
                )}
              </Button>
            )
          })}
        </div>
      </header>

      {countsError && !notConnected && !error && (
        <Alert variant="destructive">
          <AlertDescription>
            Could not load list counts. {countsError}
          </AlertDescription>
          <AlertAction>
            <Button
              type="button"
              variant="secondary"
              onClick={() => void loadCounts()}
            >
              Try again
            </Button>
          </AlertAction>
        </Alert>
      )}

      {loading ? (
        <ul aria-busy="true" aria-label="Loading your list">
          {Array.from({length: 5}, (_, index) => (
            <li
              key={index}
              className="flex items-center gap-4 border-b border-border py-4 first:border-t"
            >
              <Skeleton className="h-22 w-16 shrink-0 animate-pulse" />
              <div className="flex min-w-0 flex-1 flex-col gap-3">
                <Skeleton className="h-4 w-2/3 animate-pulse" />
                <Skeleton className="h-3 w-1/4 animate-pulse" />
                <Skeleton className="h-2 w-full animate-pulse" />
              </div>
              <Skeleton className="h-11 w-16 animate-pulse" />
            </li>
          ))}
        </ul>
      ) : notConnected ? (
        <section className="border border-border bg-card p-5">
          <h3 className="font-medium">Connect AniList to see your list</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Sign in with AniList from Settings, then return here to load your
            anime list.
          </p>
          <Button
            type="button"
            variant="secondary"
            className="mt-4"
            onClick={() => goToSettings('settings')}
          >
            Open Settings
          </Button>
        </section>
      ) : error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
          <AlertAction>
            <Button
              type="button"
              variant="secondary"
              onClick={() => void loadList()}
            >
              Try again
            </Button>
          </AlertAction>
        </Alert>
      ) : entries.length === 0 ? (
        <p className="border border-dashed border-border/60 p-8 text-sm text-muted-foreground">
          {emptyCopy}
        </p>
      ) : (
        <section aria-label="Anime list entries">
          <ul className="border-t border-border">
            {entries.map((entry) => {
              const title = entry.titleEnglish || entry.titleRomaji
              const hasTotal = entry.totalEpisodes > 0
              const total = hasTotal ? entry.totalEpisodes : '?'
              const progressValue = hasTotal
                ? Math.min(
                    100,
                    Math.max(0, (entry.progress / entry.totalEpisodes) * 100),
                  )
                : 0
              return (
                <li
                  key={entry.mediaId}
                  className="flex items-center gap-4 border-b border-border py-4"
                >
                  {entry.coverImage ? (
                    <img
                      src={entry.coverImage}
                      alt=""
                      width={64}
                      height={88}
                      className="h-22 w-16 shrink-0 object-cover"
                    />
                  ) : (
                    <span
                      className="h-22 w-16 shrink-0 bg-muted"
                      aria-hidden="true"
                    />
                  )}
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-baseline gap-x-2 gap-y-1">
                      <p className="min-w-0 truncate font-medium">{title}</p>
                      {entry.mediaStatus && (
                        <p className="text-xs text-muted-foreground">
                          {mediaStatusLabel(entry.mediaStatus)}
                        </p>
                      )}
                    </div>
                    <div className="mt-3 flex items-baseline justify-between gap-3">
                      <p className="text-sm text-muted-foreground">
                        Episode{' '}
                        <span className="tabular-nums text-foreground">
                          {entry.progress}
                        </span>{' '}
                        <span className="tabular-nums">/ {total}</span>
                      </p>
                      {hasTotal && (
                        <p className="text-xs tabular-nums text-muted-foreground">
                          {Math.round(progressValue)}%
                        </p>
                      )}
                    </div>
                    <Progress
                      value={progressValue}
                      aria-label={`${title}: episode ${entry.progress} of ${total}`}
                      className="mt-2"
                    />
                  </div>
                  <Button
                    type="button"
                    variant="ghost"
                    className="self-stretch"
                    onClick={() => setEditingEntry(entry)}
                  >
                    Edit
                  </Button>
                </li>
              )
            })}
          </ul>
        </section>
      )}

      {editingEntry && (
        <WatchingEditSheet
          entry={editingEntry}
          saving={savingEntry}
          onClose={() => setEditingEntry(null)}
          onSave={(input) => void saveEntryWithState(input)}
        />
      )}

      {searchOpen && (
        <WatchingSearchDialog
          notice={notice}
          onClose={() => setSearchOpen(false)}
        />
      )}
    </section>
  )
}
