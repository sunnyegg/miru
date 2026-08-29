import {useEffect, useRef, useState} from 'react'
import {GetAnime} from '../../wailsjs/go/main/App'
import {IconCalendar, IconCheck, IconChevronDown, IconPlay} from './Icons'
import {anilistExtraLargeCover} from '../lib/anilistImage'
import {sanitizeAnilistSynopsis} from '../lib/anilistDescription'
import {dayFormatter, scheduleTitle, timeFormatter} from '../lib/calendar'
import {errorMessage} from '../lib/format'
import type {AiringScheduleView, AnimeView} from '../lib/types'
import {useWatchingStore, type QuickAddStatus} from '../stores/watchingStore'
import {Alert, AlertDescription} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'
import {Dialog} from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {Skeleton} from '@/components/ui/skeleton'

type Props = {
  schedule: AiringScheduleView | null
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
  if (!status) {
    return ''
  }
  return mediaStatusLabels[status] ?? status
}

function episodeCountLabel(totalEpisodes: number): string {
  if (totalEpisodes <= 0) {
    return 'Unknown episode count'
  }
  return `${totalEpisodes} episodes`
}

export function AiringScheduleDialog({schedule, notice, onClose}: Props) {
  const saveListStatus = useWatchingStore((state) => state.setListStatus)
  const [anime, setAnime] = useState<AnimeView | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [posterExpanded, setPosterExpanded] = useState(false)
  const [activeListStatus, setActiveListStatus] = useState('')
  const [savingStatus, setSavingStatus] = useState<QuickAddStatus | null>(null)
  const saveRequestRef = useRef(0)

  useEffect(() => {
    if (!schedule) {
      saveRequestRef.current += 1
      setAnime(null)
      setError('')
      setLoading(false)
      setPosterExpanded(false)
      setActiveListStatus('')
      setSavingStatus(null)
      return
    }

    saveRequestRef.current += 1
    setPosterExpanded(false)
    setActiveListStatus('')
    setSavingStatus(null)

    let cancelled = false

    const mediaId = schedule.mediaId

    async function loadAnimeDetails() {
      setLoading(true)
      setError('')
      setAnime(null)
      try {
        const details = await GetAnime(mediaId)
        if (!cancelled) {
          setAnime(details)
          setActiveListStatus(details.listStatus)
        }
      } catch (err) {
        if (!cancelled) {
          setError(errorMessage(err))
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    void loadAnimeDetails()

    return () => {
      cancelled = true
    }
  }, [schedule])

  useEffect(() => {
    if (!posterExpanded) {
      return
    }
    function onKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        event.stopPropagation()
        setPosterExpanded(false)
      }
    }
    document.addEventListener('keydown', onKeyDown, true)
    return () => document.removeEventListener('keydown', onKeyDown, true)
  }, [posterExpanded])

  async function updateListStatus(status: QuickAddStatus) {
    if (!schedule || savingStatus !== null || activeListStatus === status) {
      return
    }
    const request = saveRequestRef.current + 1
    saveRequestRef.current = request
    setSavingStatus(status)
    try {
      const saved = await saveListStatus(
        schedule.mediaId,
        status,
        status === 'COMPLETED' ? (anime?.totalEpisodes ?? 0) : 0,
        notice,
      )
      if (saved && saveRequestRef.current === request) {
        setActiveListStatus(status)
        setAnime((current) =>
          current ? {...current, listStatus: status} : current,
        )
      }
    } finally {
      if (saveRequestRef.current === request) {
        setSavingStatus(null)
      }
    }
  }

  if (!schedule) {
    return null
  }

  const airingDate = new Date(schedule.airingAt * 1000)
  const title = scheduleTitle(schedule)
  const romajiTitle = schedule.titleRomaji.trim()
  const showRomaji = romajiTitle.length > 0 && romajiTitle !== title
  const coverImage =
    anime?.coverImage || anilistExtraLargeCover(schedule.coverImage)
  const activeListLabel = listStatusLabels[activeListStatus]
  const listButtonLabel = savingStatus
    ? 'Updating…'
    : activeListLabel
      ? `List: ${activeListLabel}`
      : 'Add to list'

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
            aria-labelledby="airing-detail-title"
          >
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div className="min-w-0 flex-1">
                <Dialog.Title
                  id="airing-detail-title"
                  className="text-xl font-semibold"
                >
                  {title}
                </Dialog.Title>
                {showRomaji && (
                  <p className="mt-1 text-base text-muted-foreground">
                    {romajiTitle}
                  </p>
                )}
              </div>
              <div className="flex shrink-0 items-center gap-2">
                <DropdownMenu disabled={loading || savingStatus !== null}>
                  <DropdownMenuTrigger
                    render={
                      <Button
                        type="button"
                        variant="secondary"
                        disabled={loading || savingStatus !== null}
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
              {coverImage ? (
                <button
                  type="button"
                  className="mx-auto shrink-0 cursor-zoom-in sm:mx-0"
                  aria-label={`View ${title} cover art larger`}
                  onClick={() => setPosterExpanded(true)}
                >
                  <img
                    src={coverImage}
                    alt=""
                    width={224}
                    height={336}
                    className="aspect-[2/3] w-56 object-cover"
                  />
                </button>
              ) : (
                <span
                  className="mx-auto aspect-[2/3] w-56 shrink-0 bg-muted sm:mx-0"
                  aria-hidden="true"
                />
              )}

              <div className="flex min-w-0 flex-1 flex-col gap-4 text-base">
                <dl className="grid grid-cols-1 gap-x-6 gap-y-3 sm:grid-cols-2">
                  <div>
                    <dt className="text-sm text-muted-foreground">Airing</dt>
                    <dd className="font-medium">
                      {dayFormatter.format(airingDate)} ·{' '}
                      {timeFormatter.format(airingDate)}
                    </dd>
                  </div>
                  <div>
                    <dt className="text-sm text-muted-foreground">Episode</dt>
                    <dd className="font-medium">{schedule.episode}</dd>
                  </div>
                  {loading ? (
                    <>
                      <Skeleton className="h-10 w-full animate-pulse" />
                      <Skeleton className="h-10 w-full animate-pulse" />
                    </>
                  ) : error ? null : anime ? (
                    <>
                      {anime.status && (
                        <div>
                          <dt className="text-sm text-muted-foreground">
                            Status
                          </dt>
                          <dd className="font-medium">
                            {mediaStatusLabel(anime.status)}
                          </dd>
                        </div>
                      )}
                      <div>
                        <dt className="text-sm text-muted-foreground">
                          Episodes
                        </dt>
                        <dd className="font-medium">
                          {episodeCountLabel(anime.totalEpisodes)}
                        </dd>
                      </div>
                    </>
                  ) : null}
                </dl>
                {error ? (
                  <Alert variant="destructive">
                    <AlertDescription>{error}</AlertDescription>
                  </Alert>
                ) : anime?.synopsis.trim() ? (
                  <div>
                    <p className="text-sm text-muted-foreground">Synopsis</p>
                    <div
                      className="mt-1 max-h-72 overflow-y-auto text-foreground/90 [&_a]:text-accent [&_a]:underline"
                      dangerouslySetInnerHTML={{
                        __html: sanitizeAnilistSynopsis(anime.synopsis),
                      }}
                    />
                  </div>
                ) : null}
              </div>
            </div>
          </Dialog.Panel>
        </Dialog.Viewport>
        {posterExpanded && coverImage && (
          <div
            className="fixed inset-0 z-[60] flex items-center justify-center bg-bezel/90 p-6"
            role="dialog"
            aria-modal="true"
            aria-label={`${title} cover art`}
            onClick={() => setPosterExpanded(false)}
          >
            <img
              src={coverImage}
              alt=""
              className="max-h-[calc(100vh-3rem)] max-w-full object-contain"
            />
          </div>
        )}
      </Dialog.Portal>
    </Dialog.Root>
  )
}
