import {useState} from 'react'
import {
  IconCalendar,
  IconCheck,
  IconChevronDown,
  IconPlay,
  IconSearch,
} from './Icons'
import {useWatchingStore, type QuickAddStatus} from '../stores/watchingStore'
import {Button} from '@/components/ui/button'
import {Dialog} from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {Input} from '@/components/ui/input'

type Props = {
  notice: (msg: string, isError?: boolean) => void
  onClose: () => void
}

const listStatusLabels: Record<string, string> = {
  CURRENT: 'Watching',
  COMPLETED: 'Completed',
  PLANNING: 'Planning',
  PAUSED: 'Paused',
  DROPPED: 'Dropped',
  REPEATING: 'Repeating',
}

const quickAddActions: {
  status: QuickAddStatus
  label: string
  Icon: typeof IconPlay
}[] = [
  {status: 'CURRENT', label: 'Add to Watching', Icon: IconPlay},
  {status: 'PLANNING', label: 'Add to Planning', Icon: IconCalendar},
  {status: 'COMPLETED', label: 'Add to Completed', Icon: IconCheck},
]

type BusyAction = {
  mediaId: number
}

export function WatchingSearchDialog({notice, onClose}: Props) {
  const searchQuery = useWatchingStore((state) => state.searchQuery)
  const searchResults = useWatchingStore((state) => state.searchResults)
  const searching = useWatchingStore((state) => state.searching)
  const searchError = useWatchingStore((state) => state.searchError)
  const setSearchQuery = useWatchingStore((state) => state.setSearchQuery)
  const clearSearch = useWatchingStore((state) => state.clearSearch)
  const searchAnime = useWatchingStore((state) => state.searchAnime)
  const setListStatus = useWatchingStore((state) => state.setListStatus)
  const [busyAction, setBusyAction] = useState<BusyAction | null>(null)

  const canClearSearch =
    searchQuery.trim() !== '' || searchResults.length > 0 || searchError !== ''

  async function setListStatusWithBusy(
    mediaId: number,
    status: QuickAddStatus,
    totalEpisodes: number,
  ) {
    if (busyAction !== null) {
      return
    }
    setBusyAction({mediaId})
    try {
      await setListStatus(mediaId, status, totalEpisodes, notice)
    } finally {
      setBusyAction(null)
    }
  }

  return (
    <Dialog.Root
      open
      onOpenChange={(open) => {
        if (!open) {
          onClose()
        }
      }}
    >
      <Dialog.Portal>
        <Dialog.Backdrop />
        <Dialog.Viewport>
          <Dialog.Panel
            className="max-w-4xl"
            aria-labelledby="watching-search-title"
          >
            <div className="flex items-start justify-between gap-4">
              <div className="min-w-0">
                <div className="flex items-center gap-2">
                  <IconSearch className="size-4 text-muted-foreground" />
                  <Dialog.Title id="watching-search-title">
                    Find a title
                  </Dialog.Title>
                </div>
                <Dialog.Description>
                  Search AniList, then add a title to one of your lists.
                </Dialog.Description>
              </div>
              <Button type="button" variant="ghost" onClick={onClose}>
                Close
              </Button>
            </div>

            <form
              className="mt-4 flex flex-wrap gap-2"
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
                placeholder="Search AniList by title"
                autoFocus
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
              <section
                className="mt-4 border-t border-border pt-4"
                aria-label="AniList search results"
              >
                <p className="text-sm tabular-nums text-muted-foreground">
                  {searchResults.length} results
                </p>
                <ul className="mt-3 grid grid-cols-[repeat(auto-fill,minmax(min(11rem,100%),1fr))] gap-4">
                  {searchResults.map((anime) => {
                    const title = anime.titleEnglish || anime.titleRomaji
                    const listCaption = anime.listStatus
                      ? (listStatusLabels[anime.listStatus] ?? anime.listStatus)
                      : 'Not on your list'
                    return (
                      <li key={anime.id}>
                        <div className="relative aspect-2/3 w-full bg-muted">
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
                            className="absolute inset-x-0 bottom-0 min-h-[40%] bg-linear-to-t from-background via-background/95 to-transparent"
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
                          <div className="absolute top-1.5 right-1.5">
                            <DropdownMenu disabled={busyAction !== null}>
                              <DropdownMenuTrigger
                                render={
                                  <Button
                                    type="button"
                                    variant="secondary"
                                    className="h-11 min-h-11 bg-background/90 px-3 text-xs backdrop-blur-sm hover:bg-background"
                                    disabled={busyAction !== null}
                                    aria-busy={busyAction?.mediaId === anime.id}
                                  />
                                }
                              >
                                {busyAction?.mediaId === anime.id
                                  ? 'Updating…'
                                  : 'Add to list'}
                                <IconChevronDown className="size-3" />
                              </DropdownMenuTrigger>
                              <DropdownMenuContent
                                aria-label={`Add ${title} to list`}
                              >
                                <DropdownMenuRadioGroup
                                  value={anime.listStatus}
                                  disabled={busyAction !== null}
                                  onValueChange={(value) =>
                                    void setListStatusWithBusy(
                                      anime.id,
                                      value as QuickAddStatus,
                                      value === 'COMPLETED'
                                        ? anime.totalEpisodes
                                        : 0,
                                    )
                                  }
                                >
                                  {quickAddActions.map((action) => {
                                    const selected =
                                      anime.listStatus === action.status
                                    return (
                                      <DropdownMenuRadioItem
                                        key={action.status}
                                        value={action.status}
                                        closeOnClick
                                        disabled={
                                          busyAction !== null || selected
                                        }
                                      >
                                        <action.Icon className="size-4" />
                                        <span className="flex-1">
                                          {action.label}
                                        </span>
                                        <span
                                          aria-hidden="true"
                                          className={
                                            selected
                                              ? 'size-1.5 shrink-0 bg-accent'
                                              : 'size-1.5 shrink-0'
                                          }
                                        />
                                      </DropdownMenuRadioItem>
                                    )
                                  })}
                                </DropdownMenuRadioGroup>
                              </DropdownMenuContent>
                            </DropdownMenu>
                          </div>
                        </div>
                      </li>
                    )
                  })}
                </ul>
              </section>
            )}
          </Dialog.Panel>
        </Dialog.Viewport>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
