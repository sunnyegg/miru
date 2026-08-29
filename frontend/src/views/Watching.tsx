import {useEffect, useState} from 'react'
import {IconCalendar, IconCheck, IconPlay} from '../components/Icons'
import {WatchingEditSheet} from '../components/WatchingEditSheet'
import type {AnimeListEntryInput} from '../lib/types'
import {useNavigationStore} from '../stores/navigationStore'
import {
  useWatchingStore,
  type ListFilter,
  type QuickAddStatus,
} from '../stores/watchingStore'
import {Alert, AlertAction, AlertDescription} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'
import {Card} from '@/components/ui/card'
import {Input} from '@/components/ui/input'
import {Skeleton} from '@/components/ui/skeleton'
import {Tooltip, TooltipContent, TooltipTrigger} from '@/components/ui/tooltip'
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

const quickAddActions: {
  status: QuickAddStatus
  label: string
  Icon: typeof IconPlay
}[] = [
  {status: 'CURRENT', label: 'Add to Watching', Icon: IconPlay},
  {status: 'COMPLETED', label: 'Add to Completed', Icon: IconCheck},
  {status: 'PLANNING', label: 'Add to Planning', Icon: IconCalendar},
]

type BusyAction = {
  mediaId: number
  status: QuickAddStatus
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
  const loading = useWatchingStore((state) => state.loading)
  const notConnected = useWatchingStore((state) => state.notConnected)
  const error = useWatchingStore((state) => state.error)
  const searchQuery = useWatchingStore((state) => state.searchQuery)
  const searchResults = useWatchingStore((state) => state.searchResults)
  const searching = useWatchingStore((state) => state.searching)
  const searchError = useWatchingStore((state) => state.searchError)
  const setSearchQuery = useWatchingStore((state) => state.setSearchQuery)
  const clearSearch = useWatchingStore((state) => state.clearSearch)
  const selectFilter = useWatchingStore((state) => state.selectFilter)
  const loadList = useWatchingStore((state) => state.loadList)
  const searchAnime = useWatchingStore((state) => state.searchAnime)
  const setListStatus = useWatchingStore((state) => state.setListStatus)
  const saveEntry = useWatchingStore((state) => state.saveEntry)

  const [busyAction, setBusyAction] = useState<BusyAction | null>(null)
  const [editingEntry, setEditingEntry] = useState<
    (typeof entries)[number] | null
  >(null)
  const [savingEntry, setSavingEntry] = useState(false)

  useEffect(() => {
    void loadList()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function setListStatusWithBusy(
    mediaId: number,
    status: QuickAddStatus,
    totalEpisodes: number,
  ) {
    setBusyAction({mediaId, status})
    try {
      await setListStatus(mediaId, status, totalEpisodes, notice)
    } finally {
      setBusyAction(null)
    }
  }

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
  const emptyCopy = `Nothing on your ${filterLabel} list. Switch filter, or search AniList above to add a title.`
  const canClearSearch =
    searchQuery.trim() !== '' || searchResults.length > 0 || searchError !== ''

  return (
    <section className="flex h-full flex-col gap-6">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-2xl font-semibold">{filterLabel}</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Your AniList anime list and watch progress.
          </p>
        </div>
        <div className="flex flex-wrap gap-1">
          {listFilters.map((filter) => (
            <Button
              key={filter.value}
              type="button"
              variant={listFilter === filter.value ? 'muted' : 'ghost'}
              onClick={() => void selectFilter(filter.value)}
            >
              {filter.label}
            </Button>
          ))}
        </div>
      </header>

      {!notConnected && (
        <Card className="border border-border p-4">
          <h3 className="text-base font-medium">Search AniList</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Find anime and add it to your list.
          </p>
          <form
            className="mt-3 flex flex-wrap gap-2"
            onSubmit={(event) => {
              event.preventDefault()
              void searchAnime()
            }}
          >
            <label className="sr-only" htmlFor="watching-anilist-search">
              Search AniList
            </label>
            <Input
              id="watching-anilist-search"
              value={searchQuery}
              onChange={(event) => setSearchQuery(event.target.value)}
              className="min-w-0 flex-1 basis-56"
              placeholder="Anime title"
            />
            <Button
              type="submit"
              variant="secondary"
              disabled={searching || !searchQuery.trim()}
            >
              {searching ? 'Searching…' : 'Search'}
            </Button>
            <Button
              type="button"
              variant="ghost"
              onClick={() => clearSearch()}
              disabled={searching || !canClearSearch}
            >
              Clear
            </Button>
          </form>
          {searchError && (
            <p className="mt-2 text-sm text-destructive">{searchError}</p>
          )}
          {searchResults.length > 0 && (
            <ul className="mt-4 grid grid-cols-[repeat(auto-fill,minmax(min(13rem,100%),1fr))] gap-5">
              {searchResults.map((anime) => {
                const title = anime.titleEnglish || anime.titleRomaji
                const listCaption = anime.listStatus
                  ? listStatusLabel(anime.listStatus)
                  : 'Not on your list'
                return (
                  <li key={anime.id}>
                    <div className="flex flex-col gap-2">
                      <div className="relative aspect-[2/3] w-full bg-muted">
                        {anime.coverImage ? (
                          <img
                            src={anime.coverImage}
                            alt=""
                            referrerPolicy="no-referrer"
                            className="size-full object-cover"
                          />
                        ) : (
                          <span className="flex size-full items-end p-3 text-xs text-muted-foreground">
                            {title}
                          </span>
                        )}
                        <div
                          className="absolute inset-x-0 bottom-0 min-h-[40%] bg-gradient-to-t from-background via-background/95 to-transparent"
                          aria-hidden="true"
                        />
                        <div className="absolute inset-x-0 bottom-0 px-2 pb-2">
                          <p className="truncate text-sm font-semibold text-foreground [text-shadow:0_1px_3px_rgba(0,0,0,0.85)]">
                            {title}
                          </p>
                          <p className="mt-0.5 truncate text-xs text-foreground/80 [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                            {listCaption}
                          </p>
                        </div>
                        <div className="absolute top-1.5 right-1.5 flex flex-col gap-0.5">
                          {quickAddActions.map((action) => {
                            const isCurrentStatus =
                              anime.listStatus === action.status
                            const isBusy =
                              busyAction?.mediaId === anime.id &&
                              busyAction.status === action.status
                            const isDisabled =
                              isCurrentStatus ||
                              (busyAction !== null && !isBusy)
                            return (
                              <Tooltip key={action.status}>
                                <TooltipTrigger
                                  render={<span className="inline-flex" />}
                                >
                                  <Button
                                    type="button"
                                    variant="secondary"
                                    className={cn(
                                      'size-8 min-h-8 min-w-8 bg-background/90 backdrop-blur-sm hover:bg-background',
                                      isCurrentStatus &&
                                        'bg-accent text-accent-foreground hover:bg-accent',
                                    )}
                                    aria-label={action.label}
                                    aria-pressed={isCurrentStatus}
                                    aria-busy={isBusy}
                                    disabled={isDisabled}
                                    onClick={() =>
                                      void setListStatusWithBusy(
                                        anime.id,
                                        action.status,
                                        action.status === 'COMPLETED'
                                          ? anime.totalEpisodes
                                          : 0,
                                      )
                                    }
                                  >
                                    <action.Icon className="size-3" />
                                  </Button>
                                </TooltipTrigger>
                                <TooltipContent side="left">
                                  {action.label}
                                </TooltipContent>
                              </Tooltip>
                            )
                          })}
                        </div>
                      </div>
                    </div>
                  </li>
                )
              })}
            </ul>
          )}
        </Card>
      )}

      {loading ? (
        <ul
          className="grid gap-3 lg:grid-cols-2"
          aria-busy="true"
          aria-label="Loading your list"
        >
          {Array.from({length: 4}, (_, i) => (
            <li key={i}>
              <div className="flex flex-row items-center gap-4 bg-card p-3">
                <Skeleton className="h-16 w-12 shrink-0 animate-pulse" />
                <div className="flex flex-1 flex-col gap-2">
                  <Skeleton className="h-4 w-2/3 animate-pulse" />
                  <Skeleton className="h-3 w-1/3 animate-pulse" />
                </div>
              </div>
            </li>
          ))}
        </ul>
      ) : notConnected ? (
        <Card>
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
        </Card>
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
        <p className="border border-dashed border-border/40 p-8 text-sm text-muted-foreground">
          {emptyCopy}
        </p>
      ) : (
        <ul className="grid gap-3 lg:grid-cols-2">
          {entries.map((entry) => {
            const title = entry.titleEnglish || entry.titleRomaji
            const total = entry.totalEpisodes > 0 ? entry.totalEpisodes : '?'
            return (
              <li key={entry.mediaId}>
                <Card className="flex-row items-center gap-4 p-3">
                  {entry.coverImage ? (
                    <img
                      src={entry.coverImage}
                      alt=""
                      width={48}
                      height={64}
                      className="h-16 w-12 object-cover"
                    />
                  ) : (
                    <span className="h-16 w-12 bg-muted" aria-hidden="true" />
                  )}
                  <div className="min-w-0 flex-1">
                    <p className="truncate font-medium">{title}</p>
                    <p className="mt-1 text-sm text-muted-foreground">
                      Episode {entry.progress} / {total}
                    </p>
                    {entry.mediaStatus && (
                      <p className="mt-1 text-xs text-muted-foreground">
                        {mediaStatusLabel(entry.mediaStatus)}
                      </p>
                    )}
                  </div>
                  <Button
                    type="button"
                    variant="ghost"
                    onClick={() => setEditingEntry(entry)}
                  >
                    Edit
                  </Button>
                </Card>
              </li>
            )
          })}
        </ul>
      )}

      {editingEntry && (
        <WatchingEditSheet
          entry={editingEntry}
          saving={savingEntry}
          onClose={() => setEditingEntry(null)}
          onSave={(input) => void saveEntryWithState(input)}
        />
      )}
    </section>
  )
}
