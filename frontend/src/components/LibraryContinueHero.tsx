import {useMemo} from 'react'
import {libraryHeroBackgroundImage} from '../lib/anilistImage'
import type {ShowGroup} from '../lib/groupEpisodes'
import {
  buildWatchingShowItems,
  pickContinueHeroKey,
  torrentSearchQuery,
  watchingPosterCaption,
  type WatchingShowItem,
} from '../lib/libraryWatching'
import type {EpisodeView, PlaybackEvent, WatchingEntryView} from '../lib/types'
import {Button} from '@/components/ui/button'

type Props = {
  entries: WatchingEntryView[]
  localShows: ShowGroup[]
  libraryEpisodes: EpisodeView[]
  playing: PlaybackEvent | null
  lastPlayback: PlaybackEvent | null
  playingShowKey: string | null
  onOpenShow: (localShowKey: string) => void
  onFindTorrent: (query: string) => void
}

function heroProgressCaption(
  show: ShowGroup,
  item: WatchingShowItem | null,
  lastPlayback: PlaybackEvent | null,
): string {
  if (lastPlayback) {
    const episode = show.episodes.find(
      (entry) => entry.id === lastPlayback.episodeId,
    )
    if (episode && episode.episodeNumber > 0) {
      const percent = Number.isFinite(lastPlayback.percent)
        ? Math.min(100, Math.max(0, Math.round(lastPlayback.percent)))
        : 0
      if (percent > 0) {
        return `Episode ${episode.episodeNumber} · ${percent}% watched`
      }
      return `Episode ${episode.episodeNumber}`
    }
  }
  if (item) {
    return watchingPosterCaption(item).text
  }
  if (show.totalEpisodes > 0) {
    const remaining = show.totalEpisodes - show.progress
    if (remaining > 0) {
      return `Episode ${show.progress} / ${show.totalEpisodes} · ${remaining} left`
    }
    return `Episode ${show.progress} / ${show.totalEpisodes}`
  }
  return `${show.progress} watched`
}

function findWatchingItem(
  items: WatchingShowItem[],
  key: string,
): WatchingShowItem | null {
  return items.find((item) => item.key === key) ?? null
}

export function LibraryContinueHero({
  entries,
  localShows,
  libraryEpisodes,
  playing,
  lastPlayback,
  playingShowKey,
  onOpenShow,
  onFindTorrent,
}: Props) {
  const heroKey = useMemo(
    () =>
      pickContinueHeroKey(
        entries,
        localShows,
        playingShowKey,
        lastPlayback?.episodeId ?? null,
        libraryEpisodes,
      ),
    [
      entries,
      localShows,
      playingShowKey,
      lastPlayback?.episodeId,
      libraryEpisodes,
    ],
  )

  const watchingItems = useMemo(
    () => buildWatchingShowItems(entries, localShows),
    [entries, localShows],
  )

  if (!heroKey) {
    return null
  }

  const show = localShows.find((entry) => entry.key === heroKey)
  const watchingItem = findWatchingItem(watchingItems, heroKey)
  const title = show?.title ?? watchingItem?.title ?? 'Continue watching'
  const heroBackground = libraryHeroBackgroundImage(watchingItem, show)
  const isPlaying = playingShowKey === heroKey && playing !== null
  const progressCaption = show
    ? heroProgressCaption(show, watchingItem, lastPlayback)
    : watchingItem
      ? watchingPosterCaption(watchingItem).text
      : ''

  function handleContinue() {
    if (watchingItem?.hasLocalFiles && watchingItem.localShowKey) {
      onOpenShow(watchingItem.localShowKey)
      return
    }
    if (show) {
      onOpenShow(show.key)
    }
  }

  function handleFindTorrent() {
    if (!watchingItem) {
      return
    }
    const episodeNumber =
      watchingItem.newEpisodeNumber ?? watchingItem.progress + 1
    onFindTorrent(torrentSearchQuery(watchingItem.title, episodeNumber))
  }

  const showFindTorrent =
    watchingItem !== null &&
    !watchingItem.hasLocalFiles &&
    watchingItem.newEpisodeNumber !== null

  return (
    <section
      className="relative mb-8 overflow-hidden"
      aria-label="Continue watching"
    >
      <div className="relative flex min-h-44 flex-col justify-end sm:min-h-52">
        {heroBackground ? (
          <img
            src={heroBackground}
            alt=""
            referrerPolicy="no-referrer"
            className="absolute inset-0 size-full object-cover object-top opacity-50"
          />
        ) : (
          <div className="absolute inset-0 bg-muted" />
        )}
        <div className="absolute inset-0 bg-gradient-to-r from-background via-background/85 to-background/40" />
        <div className="relative px-1 pb-1 pt-16 sm:pt-20">
          <p className="text-xs font-medium tracking-wide text-muted-foreground uppercase">
            {isPlaying ? 'Now playing' : 'Continue watching'}
          </p>
          <h3 className="mt-1 max-w-xl text-2xl font-semibold tracking-tight sm:text-3xl">
            {title}
          </h3>
          <p className="mt-1 text-sm text-muted-foreground">
            {progressCaption}
          </p>
          <div className="mt-4 flex flex-wrap gap-2">
            <Button type="button" onClick={handleContinue}>
              {isPlaying ? 'Open show' : 'Continue'}
            </Button>
            {showFindTorrent && (
              <Button
                type="button"
                variant="secondary"
                onClick={handleFindTorrent}
              >
                Find torrent
              </Button>
            )}
          </div>
        </div>
      </div>
    </section>
  )
}
