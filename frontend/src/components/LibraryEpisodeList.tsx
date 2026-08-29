import {episodeSlots, type ShowGroup} from '../lib/groupEpisodes'
import {torrentSearchQuery} from '../lib/libraryWatching'
import type {PlaybackEvent} from '../lib/types'
import {Button} from '@/components/ui/button'
import {cn} from '@/lib/utils'

type Props = {
  show: ShowGroup
  playing: PlaybackEvent | null
  lastPlayback: PlaybackEvent | null
  busyId: number | null
  unmatchingEpisodeId: number | null
  episodeThumbnails: Record<number, string>
  onPlay: (episodeId: number) => void
  onUnmatch?: (episodeId: number) => void
  onFindTorrent?: (query: string) => void
}

const rowClassName = 'flex min-h-[4.5rem] items-center gap-3 bg-card/60 px-4 py-1'
const episodeActionButtonClassName = 'w-32 shrink-0 justify-center'
const episodeNumberClassName = 'w-10 shrink-0 tabular-nums text-sm'
const titleClassName = 'block truncate text-base font-medium'
const subtitleClassName = 'block truncate text-sm'

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

function episodeThumbnailUrl(
  episodeNumber: number,
  episodeThumbnails: Record<number, string>,
  showCover: string,
): string {
  if (episodeNumber > 0 && episodeThumbnails[episodeNumber]) {
    return episodeThumbnails[episodeNumber]
  }
  return showCover
}

function EpisodeThumbnail({
  imageUrl,
  dimmed = false,
}: {
  imageUrl: string
  dimmed?: boolean
}) {
  if (!imageUrl) {
    return (
      <span
        className="aspect-video h-16 w-28 shrink-0 bg-muted"
        aria-hidden="true"
      />
    )
  }

  return (
    <img
      src={imageUrl}
      alt=""
      referrerPolicy="no-referrer"
      className={cn(
        'aspect-video h-16 w-28 shrink-0 bg-muted object-cover',
        dimmed && 'opacity-50',
      )}
      aria-hidden="true"
    />
  )
}

function episodePlaybackState(
  episodeId: number,
  playing: PlaybackEvent | null,
  lastPlayback: PlaybackEvent | null,
): {highlighted: boolean; isPlaying: boolean; percent: number} {
  if (playing?.episodeId === episodeId) {
    const percent = Number.isFinite(playing.percent)
      ? Math.min(100, Math.max(0, playing.percent))
      : 0
    return {highlighted: true, isPlaying: true, percent}
  }
  if (lastPlayback?.episodeId === episodeId) {
    const percent = Number.isFinite(lastPlayback.percent)
      ? Math.min(100, Math.max(0, lastPlayback.percent))
      : 0
    return {highlighted: true, isPlaying: false, percent}
  }
  return {highlighted: false, isPlaying: false, percent: 0}
}

export function LibraryEpisodeList({
  show,
  playing,
  lastPlayback,
  busyId,
  unmatchingEpisodeId,
  episodeThumbnails,
  onPlay,
  onUnmatch,
  onFindTorrent,
}: Props) {
  const slots = episodeSlots(show)

  function findTorrentForEpisode(episodeNumber: number) {
    if (!onFindTorrent || episodeNumber <= 0) {
      return
    }
    onFindTorrent(torrentSearchQuery(show.title, episodeNumber))
  }

  return (
    <ul className="flex flex-col gap-2.5" aria-label={`Episodes for ${show.title}`}>
      {slots.map((slot) => {
        const label = episodeLabel(slot.number, slot.file?.displayTitle)
        const subtitle = slot.kind === 'upcoming'
          ? 'Not yet aired'
          : slot.kind === 'missing'
            ? 'No file'
            : slot.file?.filePath.split(/[\\/]/).pop() || slot.file?.displayTitle || ''
        const thumbnailUrl = episodeThumbnailUrl(slot.number, episodeThumbnails, show.coverImage)

        if (slot.kind === 'upcoming') {
          return (
            <li key={slotKey(slot)}>
              <div
                aria-label={`${label}, ${subtitle}`}
                className={cn(rowClassName, 'text-muted-foreground')}
              >
                <EpisodeThumbnail imageUrl={thumbnailUrl} dimmed />
                <span className={episodeNumberClassName}>
                  {slot.number > 0 ? slot.number : '—'}
                </span>
                <span className="min-w-0 flex-1 py-2.5">
                  <span className={titleClassName}>{label}</span>
                  <span className={subtitleClassName}>{subtitle}</span>
                </span>
              </div>
            </li>
          )
        }

        if (slot.kind === 'missing' && onFindTorrent && slot.number > 0) {
          return (
            <li key={slotKey(slot)}>
              <div className={rowClassName}>
                <EpisodeThumbnail imageUrl={thumbnailUrl} dimmed />
                <span className={cn(episodeNumberClassName, 'text-muted-foreground')}>
                  {slot.number}
                </span>
                <span className="min-w-0 flex-1 py-2.5">
                  <span className={titleClassName}>{label}</span>
                  <span className={cn(subtitleClassName, 'text-muted-foreground')}>{subtitle}</span>
                </span>
                <Button
                  type="button"
                  variant="secondary"
                  className={episodeActionButtonClassName}
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
                className={cn(rowClassName, 'text-muted-foreground')}
              >
                <EpisodeThumbnail imageUrl={thumbnailUrl} dimmed />
                <span className={episodeNumberClassName}>
                  {slot.number > 0 ? slot.number : '—'}
                </span>
                <span className="min-w-0 flex-1 py-2.5">
                  <span className={titleClassName}>{label}</span>
                  <span className={subtitleClassName}>{subtitle}</span>
                </span>
              </div>
            </li>
          )
        }

        const episode = slot.file
        const isBusy = busyId === episode.id
        const isUnmatching = unmatchingEpisodeId === episode.id
        const playback = episodePlaybackState(episode.id, playing, lastPlayback)

        return (
          <li key={slotKey(slot)}>
            <div
              className={cn(
                'relative overflow-hidden border-l-2',
                rowClassName,
                playback.highlighted ? 'border-l-accent' : 'border-l-transparent',
                (isBusy || isUnmatching) && 'opacity-50',
              )}
            >
              {playback.highlighted && playback.percent > 0 && (
                <div
                  className="pointer-events-none absolute inset-x-0 bottom-0 h-2 bg-muted"
                  role="progressbar"
                  aria-label={`Playback progress for ${label}`}
                  aria-valuemin={0}
                  aria-valuemax={100}
                  aria-valuenow={Math.round(playback.percent)}
                  aria-valuetext={`${Math.round(playback.percent)}% played`}
                >
                  <div
                    className={cn(
                      'h-full bg-accent',
                      playback.isPlaying && 'transition-[width] duration-200 motion-reduce:transition-none',
                    )}
                    style={{width: `${playback.percent}%`}}
                  />
                </div>
              )}
              <EpisodeThumbnail imageUrl={thumbnailUrl} />
              <span className={cn(episodeNumberClassName, 'text-muted-foreground')}>
                {slot.number > 0 ? slot.number : '—'}
              </span>
              <span className="min-w-0 flex-1 py-2.5">
                <span className={titleClassName}>{label}</span>
                <span className={cn(subtitleClassName, 'text-muted-foreground')}>{subtitle}</span>
              </span>
              <div className="flex shrink-0 gap-2">
                {episode.bound && onUnmatch && (
                  <Button
                    type="button"
                    variant="secondary"
                    className={episodeActionButtonClassName}
                    disabled={isBusy || isUnmatching || unmatchingEpisodeId !== null}
                    onClick={() => onUnmatch(episode.id)}
                  >
                    {isUnmatching ? 'Removing…' : 'Unmatch'}
                  </Button>
                )}
                <Button
                  type="button"
                  variant="default"
                  className={episodeActionButtonClassName}
                  disabled={isBusy || isUnmatching}
                  onClick={() => onPlay(episode.id)}
                >
                  {isBusy ? 'Starting…' : 'Play'}
                </Button>
              </div>
            </div>
          </li>
        )
      })}
    </ul>
  )
}
