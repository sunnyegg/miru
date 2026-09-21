import {memo, useRef, type RefObject} from 'react'
import {episodeSlots, type ShowGroup} from '../lib/groupEpisodes'
import {torrentSearchQuery} from '../lib/libraryWatching'
import type {EpisodeView} from '../lib/types'
import {usePlaybackStore} from '../stores/playbackStore'
import {Button} from '@/components/ui/button'
import {cn} from '@/lib/utils'

type Props = {
  show: ShowGroup
  playingEpisodeId: number | null
  busyId: number | null
  unmatchingEpisodeId: number | null
  episodeThumbnails: Record<number, string>
  onPlay: (episodeId: number) => void
  onUnmatch?: (episodeId: number) => void
  onFindTorrent?: (query: string) => void
}

type RowProps = {
  episode: EpisodeView
  thumbnailUrl: string
  label: string
  subtitle: string
  isLatest: boolean
  isNextUnwatched: boolean
  nextUnwatchedRef: RefObject<HTMLLIElement | null>
  isBusy: boolean
  isUnmatching: boolean
  unmatchingBusy: boolean
  onPlay: (episodeId: number) => void
  onUnmatch?: (episodeId: number) => void
}

const rowClassName =
  'flex min-h-[4.5rem] items-center gap-3 bg-card/60 px-4 py-1'
const episodeActionButtonClassName = 'w-32 shrink-0 justify-center'
const episodeNumberClassName = 'w-10 shrink-0 tabular-nums text-sm'
const titleClassName = 'block min-w-0 truncate text-base font-medium'
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

function latestPlayedEpisodeId(
  episodes: EpisodeView[],
  playingEpisodeId: number | null,
  lastPlaybackEpisodeId: number | null,
): number | null {
  const sessionEpisodeId = playingEpisodeId ?? lastPlaybackEpisodeId
  if (
    sessionEpisodeId !== undefined &&
    episodes.some((episode) => episode.id === sessionEpisodeId)
  ) {
    return sessionEpisodeId
  }

  let latestId: number | null = null
  let latestPlayedAt = 0
  for (const episode of episodes) {
    if (!episode.lastPlayedAt) {
      continue
    }
    const playedAt = Date.parse(episode.lastPlayedAt)
    if (Number.isNaN(playedAt) || playedAt <= latestPlayedAt) {
      continue
    }
    latestPlayedAt = playedAt
    latestId = episode.id
  }
  return latestId
}

function clampPercent(percent: number): number {
  if (!Number.isFinite(percent)) {
    return 0
  }
  return Math.min(100, Math.max(0, percent))
}

const LibraryEpisodeRow = memo(function LibraryEpisodeRow({
  episode,
  thumbnailUrl,
  label,
  subtitle,
  isLatest,
  isNextUnwatched,
  nextUnwatchedRef,
  isBusy,
  isUnmatching,
  unmatchingBusy,
  onPlay,
  onUnmatch,
}: RowProps) {
  const isPlaying = usePlaybackStore(
    (state) => state.playing?.episodeId === episode.id,
  )
  const sessionPercent = usePlaybackStore(
    (state) => state.progressByEpisodeId[episode.id],
  )
  const lastPlaybackPercent = usePlaybackStore((state) =>
    state.lastPlayback?.episodeId === episode.id
      ? state.lastPlayback.percent
      : null,
  )

  let percent = clampPercent(episode.playbackPercent)
  if (lastPlaybackPercent !== null && percent === 0) {
    percent = clampPercent(lastPlaybackPercent)
  }
  if (sessionPercent !== undefined) {
    percent = clampPercent(sessionPercent)
  }
  const highlighted =
    percent > 0 ||
    Boolean(episode.lastPlayedAt) ||
    isPlaying ||
    lastPlaybackPercent !== null

  return (
    <li ref={isNextUnwatched ? nextUnwatchedRef : undefined}>
      <div
        className={cn(
          'relative overflow-hidden border-l-2',
          rowClassName,
          highlighted ? 'border-l-accent' : 'border-l-transparent',
          (isBusy || isUnmatching) && 'opacity-50',
        )}
      >
        {percent > 0 && (
          <div
            className="pointer-events-none absolute inset-x-0 bottom-0 h-2 bg-muted"
            role="progressbar"
            aria-label={`Playback progress for ${label}`}
            aria-valuemin={0}
            aria-valuemax={100}
            aria-valuenow={Math.round(percent)}
            aria-valuetext={`${Math.round(percent)}% played`}
          >
            <div
              className={cn(
                'h-full bg-accent',
                isPlaying &&
                  'transition-[width] duration-200 motion-reduce:transition-none',
              )}
              style={{width: `${percent}%`}}
            />
          </div>
        )}
        <EpisodeThumbnail imageUrl={thumbnailUrl} />
        <span className={cn(episodeNumberClassName, 'text-muted-foreground')}>
          {episode.episodeNumber > 0 ? episode.episodeNumber : '—'}
        </span>
        <span className="min-w-0 flex-1 py-2.5">
          <span className="flex min-w-0 items-center gap-2">
            <span className={titleClassName}>{label}</span>
            {isLatest && (
              <span className="shrink-0 text-xs font-medium text-accent">
                Latest watched
              </span>
            )}
          </span>
          <span className={cn(subtitleClassName, 'text-muted-foreground')}>
            {subtitle}
          </span>
        </span>
        <div className="flex shrink-0 gap-2">
          {episode.bound && onUnmatch && (
            <Button
              type="button"
              variant="secondary"
              className={episodeActionButtonClassName}
              disabled={isBusy || isUnmatching || unmatchingBusy}
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
})

export function LibraryEpisodeList({
  show,
  playingEpisodeId,
  busyId,
  unmatchingEpisodeId,
  episodeThumbnails,
  onPlay,
  onUnmatch,
  onFindTorrent,
}: Props) {
  const lastPlaybackEpisodeId = usePlaybackStore(
    (state) => state.lastPlayback?.episodeId ?? null,
  )
  const slots = episodeSlots(show)
  const latestEpisodeId = latestPlayedEpisodeId(
    show.episodes,
    playingEpisodeId,
    lastPlaybackEpisodeId,
  )
  const nextUnwatchedSlot = show.bound
    ? (slots.find((slot) => slot.number === show.progress + 1) ?? null)
    : null
  const nextUnwatchedRef = useRef<HTMLLIElement | null>(null)

  function scrollToNextUnwatched() {
    nextUnwatchedRef.current?.scrollIntoView({
      behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches
        ? 'auto'
        : 'smooth',
      block: 'start',
    })
  }

  function findTorrentForEpisode(episodeNumber: number) {
    if (!onFindTorrent || episodeNumber <= 0) {
      return
    }
    onFindTorrent(torrentSearchQuery(show.title, episodeNumber))
  }

  return (
    <>
      {nextUnwatchedSlot && (
        <div className="mb-2.5 flex justify-end">
          <Button
            type="button"
            variant="secondary"
            onClick={scrollToNextUnwatched}
          >
            Next unwatched
          </Button>
        </div>
      )}
      <ul
        className="flex flex-col gap-2.5"
        aria-label={`Episodes for ${show.title}`}
      >
        {slots.map((slot) => {
          const label = episodeLabel(slot.number, slot.file?.displayTitle)
          const subtitle =
            slot.kind === 'upcoming'
              ? 'Not yet aired'
              : slot.kind === 'missing'
                ? 'No file'
                : slot.file?.filePath.split(/[\\/]/).pop() ||
                  slot.file?.displayTitle ||
                  ''
          const thumbnailUrl = episodeThumbnailUrl(
            slot.number,
            episodeThumbnails,
            show.coverImage,
          )

          if (slot.kind === 'upcoming') {
            return (
              <li
                key={slotKey(slot)}
                ref={slot === nextUnwatchedSlot ? nextUnwatchedRef : undefined}
              >
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
              <li
                key={slotKey(slot)}
                ref={slot === nextUnwatchedSlot ? nextUnwatchedRef : undefined}
              >
                <div className={rowClassName}>
                  <EpisodeThumbnail imageUrl={thumbnailUrl} dimmed />
                  <span
                    className={cn(
                      episodeNumberClassName,
                      'text-muted-foreground',
                    )}
                  >
                    {slot.number}
                  </span>
                  <span className="min-w-0 flex-1 py-2.5">
                    <span className={titleClassName}>{label}</span>
                    <span
                      className={cn(subtitleClassName, 'text-muted-foreground')}
                    >
                      {subtitle}
                    </span>
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
              <li
                key={slotKey(slot)}
                ref={slot === nextUnwatchedSlot ? nextUnwatchedRef : undefined}
              >
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
          return (
            <LibraryEpisodeRow
              key={slotKey(slot)}
              episode={episode}
              thumbnailUrl={thumbnailUrl}
              label={label}
              subtitle={subtitle}
              isLatest={latestEpisodeId === episode.id}
              isNextUnwatched={slot === nextUnwatchedSlot}
              nextUnwatchedRef={nextUnwatchedRef}
              isBusy={busyId === episode.id}
              isUnmatching={unmatchingEpisodeId === episode.id}
              unmatchingBusy={unmatchingEpisodeId !== null}
              onPlay={onPlay}
              onUnmatch={onUnmatch}
            />
          )
        })}
      </ul>
    </>
  )
}
