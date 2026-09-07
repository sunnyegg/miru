import {useState} from 'react'
import {IconCalendar, IconCheck, IconChevronDown, IconPlay} from './Icons'
import {sanitizeAnilistSynopsis} from '../lib/anilistDescription'
import type {AnimeView} from '../lib/types'
import {useWatchingStore, type QuickAddStatus} from '../stores/watchingStore'
import {Button} from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {Dialog} from '@/components/ui/dialog'

type Props = {
  anime: AnimeView
  notice: (msg: string, isError?: boolean) => void
  onClose: () => void
  onStatusChange: (status: QuickAddStatus) => void
}

const listStatusLabels: Record<string, string> = {
  CURRENT: 'Watching',
  COMPLETED: 'Completed',
  PLANNING: 'Planning',
  PAUSED: 'Paused',
  DROPPED: 'Dropped',
  REPEATING: 'Repeating',
}

const listActions: {
  status: QuickAddStatus
  label: string
  Icon: typeof IconPlay
}[] = [
  {status: 'CURRENT', label: 'Add to Watching', Icon: IconPlay},
  {status: 'PLANNING', label: 'Add to Planning', Icon: IconCalendar},
  {status: 'COMPLETED', label: 'Add to Completed', Icon: IconCheck},
]

const mediaStatusLabels: Record<string, string> = {
  RELEASING: 'Airing',
  FINISHED: 'Finished',
  NOT_YET_RELEASED: 'Not yet aired',
  CANCELLED: 'Cancelled',
  HIATUS: 'On hiatus',
}

function mediaStatusLabel(status: string): string {
  return mediaStatusLabels[status] ?? status
}

function episodeCountLabel(totalEpisodes: number): string {
  return totalEpisodes > 0
    ? `${totalEpisodes} episodes`
    : 'Unknown episode count'
}

export function AnimeDetailDialog({
  anime,
  notice,
  onClose,
  onStatusChange,
}: Props) {
  const saveListStatus = useWatchingStore((state) => state.setListStatus)
  const [activeListStatus, setActiveListStatus] = useState(anime.listStatus)
  const [savingStatus, setSavingStatus] = useState<QuickAddStatus | null>(null)

  const title = anime.titleEnglish || anime.titleRomaji
  const romajiTitle = anime.titleRomaji.trim()
  const showRomaji = romajiTitle.length > 0 && romajiTitle !== title
  const activeListLabel = listStatusLabels[activeListStatus]
  const listButtonLabel = savingStatus
    ? 'Updating…'
    : activeListLabel
      ? `List: ${activeListLabel}`
      : 'Add to list'

  async function updateListStatus(status: QuickAddStatus) {
    if (savingStatus !== null || activeListStatus === status) {
      return
    }
    setSavingStatus(status)
    try {
      const saved = await saveListStatus(
        anime.id,
        status,
        status === 'COMPLETED' ? anime.totalEpisodes : 0,
        notice,
      )
      if (saved) {
        setActiveListStatus(status)
        onStatusChange(status)
      }
    } finally {
      setSavingStatus(null)
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
            className="max-w-4xl p-6"
            aria-labelledby="anime-detail-title"
          >
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div className="min-w-0 flex-1">
                <Dialog.Title
                  id="anime-detail-title"
                  className="text-xl font-semibold"
                >
                  {title}
                </Dialog.Title>
                {showRomaji && (
                  <Dialog.Description className="mt-1 text-base">
                    {romajiTitle}
                  </Dialog.Description>
                )}
              </div>
              <div className="flex shrink-0 items-center gap-2">
                <DropdownMenu disabled={savingStatus !== null}>
                  <DropdownMenuTrigger
                    render={
                      <Button
                        type="button"
                        variant="secondary"
                        disabled={savingStatus !== null}
                        aria-busy={savingStatus !== null}
                      />
                    }
                  >
                    {listButtonLabel}
                    <IconChevronDown className="size-4" />
                  </DropdownMenuTrigger>
                  <DropdownMenuContent aria-label="Add anime to list">
                    <DropdownMenuRadioGroup
                      value={activeListStatus}
                      disabled={savingStatus !== null}
                      onValueChange={(value) =>
                        void updateListStatus(value as QuickAddStatus)
                      }
                    >
                      {listActions.map((action) => {
                        const selected = activeListStatus === action.status
                        return (
                          <DropdownMenuRadioItem
                            key={action.status}
                            value={action.status}
                            closeOnClick
                            disabled={savingStatus !== null || selected}
                          >
                            <action.Icon className="size-4" />
                            <span className="flex-1">{action.label}</span>
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
                <Button type="button" variant="ghost" onClick={onClose}>
                  Close
                </Button>
              </div>
            </div>

            <div className="mt-6 flex flex-col gap-6 sm:flex-row">
              {anime.coverImage ? (
                <img
                  src={anime.coverImage}
                  alt=""
                  width={224}
                  height={336}
                  referrerPolicy="no-referrer"
                  className="mx-auto aspect-2/3 w-56 shrink-0 object-cover sm:mx-0"
                />
              ) : (
                <span
                  className="mx-auto aspect-2/3 w-56 shrink-0 bg-muted sm:mx-0"
                  aria-hidden="true"
                />
              )}

              <div className="flex min-w-0 flex-1 flex-col gap-4 text-base">
                <dl className="grid grid-cols-1 gap-x-6 gap-y-3 sm:grid-cols-2">
                  {anime.status && (
                    <div>
                      <dt className="text-sm text-muted-foreground">Status</dt>
                      <dd className="font-medium">
                        {mediaStatusLabel(anime.status)}
                      </dd>
                    </div>
                  )}
                  <div>
                    <dt className="text-sm text-muted-foreground">Episodes</dt>
                    <dd className="font-medium">
                      {episodeCountLabel(anime.totalEpisodes)}
                    </dd>
                  </div>
                </dl>
                {anime.synopsis.trim() && (
                  <div>
                    <p className="text-sm text-muted-foreground">Synopsis</p>
                    <div
                      className="mt-1 max-h-72 overflow-y-auto text-foreground/90 [&_a]:text-accent [&_a]:underline"
                      dangerouslySetInnerHTML={{
                        __html: sanitizeAnilistSynopsis(anime.synopsis),
                      }}
                    />
                  </div>
                )}
              </div>
            </div>
          </Dialog.Panel>
        </Dialog.Viewport>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
