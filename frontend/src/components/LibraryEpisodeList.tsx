import {episodeSlots, type ShowGroup} from '../lib/groupEpisodes'
import {torrentSearchQuery} from '../lib/libraryWatching'
import type {PlaybackEvent} from '../lib/types'
import {IconPlay} from './Icons'
import {Button} from '@/components/ui/button'

type Props = {
  show: ShowGroup
  playing: PlaybackEvent | null
  busyId: number | null
  onPlay: (episodeId: number) => void
  onFindTorrent?: (query: string) => void
}

function slotKey(slot: ReturnType<typeof episodeSlots>[number]): string {
  if (slot.file) {
    return `file-${slot.file.id}`
  }
  return `${slot.kind}-${slot.number}`
}

function episodeLabel(number: number, displayTitle?: string): string {
  if (number > 0) {
    return `Episode ${number}`
  }
  return displayTitle || 'Episode'
}

export function LibraryEpisodeList({show, playing, busyId, onPlay, onFindTorrent}: Props) {
  const slots = episodeSlots(show)

  function findTorrentForEpisode(episodeNumber: number) {
    if (!onFindTorrent || episodeNumber <= 0) {
      return
    }
    onFindTorrent(torrentSearchQuery(show.title, episodeNumber))
  }

  return (
    <ul className="flex flex-col gap-1" aria-label={`Episodes for ${show.title}`}>
      {slots.map((slot) => {
        const label = episodeLabel(slot.number, slot.file?.displayTitle)
        const subtitle = slot.kind === 'upcoming'
          ? 'Akan datang'
          : slot.kind === 'missing'
            ? 'No file'
            : slot.file?.filePath.split(/[\\/]/).pop() || slot.file?.displayTitle || ''

        if (slot.kind === 'upcoming') {
          return (
            <li key={slotKey(slot)}>
              <div
                aria-label={`${label}, ${subtitle}`}
                className="flex min-h-11 items-center gap-3 bg-card px-3 text-muted-foreground transition-colors duration-200 motion-reduce:transition-none"
              >
                <span className="w-8 shrink-0 tabular-nums text-xs">
                  {slot.number > 0 ? slot.number : '—'}
                </span>
                <span className="min-w-0 flex-1 py-2">
                  <span className="block truncate text-sm">{label}</span>
                  <span className="block truncate text-xs">{subtitle}</span>
                </span>
              </div>
            </li>
          )
        }

        if (slot.kind === 'missing' && onFindTorrent && slot.number > 0) {
          return (
            <li key={slotKey(slot)}>
              <div className="flex min-h-11 items-center gap-3 bg-card px-3 transition-colors duration-200 motion-reduce:transition-none">
                <span className="w-8 shrink-0 tabular-nums text-xs text-muted-foreground">
                  {slot.number}
                </span>
                <span className="min-w-0 flex-1 py-2">
                  <span className="block truncate text-sm">{label}</span>
                  <span className="block truncate text-xs text-muted-foreground">{subtitle}</span>
                </span>
                <Button
                  type="button"
                  variant="secondary"
                  onClick={() => findTorrentForEpisode(slot.number)}
                >
                  Find torrent
                </Button>
              </div>
            </li>
          )
        }

        if (slot.kind !== 'available' || !slot.file) {
          return (
            <li key={slotKey(slot)}>
              <div
                aria-label={`${label}, ${subtitle}`}
                className="flex min-h-11 items-center gap-3 bg-card px-3 text-muted-foreground transition-colors duration-200 motion-reduce:transition-none"
              >
                <span className="w-8 shrink-0 tabular-nums text-xs">
                  {slot.number > 0 ? slot.number : '—'}
                </span>
                <span className="min-w-0 flex-1 py-2">
                  <span className="block truncate text-sm">{label}</span>
                  <span className="block truncate text-xs">{subtitle}</span>
                </span>
              </div>
            </li>
          )
        }

        const episode = slot.file
        const isPlaying = playing?.episodeId === episode.id
        const isBusy = busyId === episode.id
        const progress = isPlaying && Number.isFinite(playing?.percent)
          ? Math.min(100, Math.max(0, playing.percent))
          : 0

        return (
          <li key={slotKey(slot)}>
            <button
              type="button"
              onClick={() => onPlay(episode.id)}
              disabled={isBusy}
              aria-label={isBusy ? `Starting ${label}` : `Play ${label}`}
              className={`relative flex min-h-11 w-full cursor-pointer items-center gap-3 overflow-hidden border-l-2 bg-card px-3 text-left transition-colors duration-200 hover:bg-muted/40 motion-reduce:transition-none ${
                isPlaying ? 'border-l-accent' : 'border-l-transparent'
              } ${isBusy ? 'cursor-not-allowed opacity-50' : ''}`}
            >
              {isPlaying && (
                <div
                  className="pointer-events-none absolute inset-x-0 bottom-0 h-2 bg-muted"
                  role="progressbar"
                  aria-label={`Playback progress for ${label}`}
                  aria-valuemin={0}
                  aria-valuemax={100}
                  aria-valuenow={Math.round(progress)}
                  aria-valuetext={`${Math.round(progress)}% played`}
                >
                  <div
                    className="h-full bg-accent transition-[width] duration-200 motion-reduce:transition-none"
                    style={{width: `${progress}%`}}
                  />
                </div>
              )}
              <span className="w-8 shrink-0 tabular-nums text-xs text-muted-foreground">
                {slot.number > 0 ? slot.number : '—'}
              </span>
              <span className="min-w-0 flex-1 pb-2 pt-2">
                <span className="block truncate text-sm font-medium">{label}</span>
                <span className="block truncate text-xs text-muted-foreground">{subtitle}</span>
              </span>
              {isPlaying && (
                <span className="shrink-0 text-accent" style={{width: 16, height: 16}}>
                  <IconPlay className="size-full" />
                </span>
              )}
              {isBusy && (
                <span className="shrink-0 text-xs text-muted-foreground">Starting…</span>
              )}
            </button>
          </li>
        )
      })}
    </ul>
  )
}
