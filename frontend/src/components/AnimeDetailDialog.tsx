import {useEffect, useRef, useState} from 'react'
import {IconCalendar, IconCheck, IconChevronDown, IconPlay} from './Icons'
import {GetAnime} from '../../wailsjs/go/main/App'
import {AnimeMetadata} from './AnimeMetadata'
import {errorMessage} from '../lib/format'
import {Alert, AlertDescription} from '@/components/ui/alert'
import {Skeleton} from '@/components/ui/skeleton'
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

export function AnimeDetailDialog({
  anime,
  notice,
  onClose,
  onStatusChange,
}: Props) {
  const saveListStatus = useWatchingStore((state) => state.setListStatus)
  const [activeListStatus, setActiveListStatus] = useState(anime.listStatus)
  const [savingStatus, setSavingStatus] = useState<QuickAddStatus | null>(null)

  const [details, setDetails] = useState<AnimeView | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const saveRequestRef = useRef(0)

  useEffect(() => {
    let cancelled = false
    saveRequestRef.current += 1
    setDetails(null)
    setActiveListStatus('')
    setSavingStatus(null)
    setLoading(true)
    setError('')
    async function loadAnimeDetails() {
      try {
        const result = await GetAnime(anime.id)
        if (!cancelled) {
          setDetails(result)
          setActiveListStatus(result.listStatus)
        }
      } catch (err) {
        if (!cancelled) setError(errorMessage(err))
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    void loadAnimeDetails()
    return () => {
      cancelled = true
      saveRequestRef.current += 1
    }
  }, [anime.id])

  const displayedAnime = details ?? anime
  const title = displayedAnime.titleEnglish || displayedAnime.titleRomaji
  const romajiTitle = displayedAnime.titleRomaji.trim()
  const showRomaji = romajiTitle.length > 0 && romajiTitle !== title
  const activeListLabel = listStatusLabels[activeListStatus]
  const listButtonLabel = savingStatus
    ? 'Updating…'
    : activeListLabel
      ? `List: ${activeListLabel}`
      : 'Add to list'

  async function updateListStatus(status: QuickAddStatus) {
    if (
      !details ||
      loading ||
      savingStatus !== null ||
      activeListStatus === status
    ) {
      return
    }
    const request = ++saveRequestRef.current
    setSavingStatus(status)
    try {
      const saved = await saveListStatus(
        anime.id,
        status,
        status === 'COMPLETED' ? details.totalEpisodes : 0,
        notice,
      )
      if (saved && saveRequestRef.current === request) {
        setActiveListStatus(status)
        onStatusChange(status)
      }
    } finally {
      if (saveRequestRef.current === request) setSavingStatus(null)
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
                <DropdownMenu
                  disabled={loading || !details || savingStatus !== null}
                >
                  <DropdownMenuTrigger
                    render={
                      <Button
                        type="button"
                        variant="secondary"
                        disabled={loading || !details || savingStatus !== null}
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
                      disabled={loading || !details || savingStatus !== null}
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
              {displayedAnime.coverImage ? (
                <img
                  src={displayedAnime.coverImage}
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
                {loading ? (
                  <Skeleton className="h-40 w-full" />
                ) : error ? (
                  <Alert variant="destructive">
                    <AlertDescription>{error}</AlertDescription>
                  </Alert>
                ) : details ? (
                  <AnimeMetadata anime={details} />
                ) : null}
              </div>
            </div>
          </Dialog.Panel>
        </Dialog.Viewport>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
