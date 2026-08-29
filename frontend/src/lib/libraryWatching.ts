import type {ShowGroup} from './groupEpisodes'
import type {EpisodeView, WatchingEntryView} from './types'

export type WatchingShowItem = {
  mediaId: number
  key: string
  title: string
  coverImage: string
  bannerImage: string
  progress: number
  totalEpisodes: number
  mediaStatus: string
  nextAiringEpisode: number
  hasLocalFiles: boolean
  localShowKey: string | null
  maxLocalEpisode: number
  newEpisodeNumber: number | null
}

function showTitle(entry: WatchingEntryView): string {
  return entry.titleEnglish || entry.titleRomaji
}

function airedLatestEpisode(entry: WatchingEntryView): number {
  if (entry.nextAiringEpisode > 0) {
    return entry.nextAiringEpisode - 1
  }
  if (entry.mediaStatus === 'FINISHED' && entry.totalEpisodes > 0) {
    return entry.totalEpisodes
  }
  return 0
}

function maxLocalEpisodeNumber(show: ShowGroup | undefined): number {
  if (!show) {
    return 0
  }
  let highest = 0
  for (const episode of show.episodes) {
    if (episode.episodeNumber > highest) {
      highest = episode.episodeNumber
    }
  }
  return highest
}

function hasLocalEpisode(
  show: ShowGroup | undefined,
  episodeNumber: number,
): boolean {
  if (!show || episodeNumber <= 0) {
    return false
  }
  return show.episodes.some(
    (episode) => episode.episodeNumber === episodeNumber,
  )
}

export function buildWatchingShowItems(
  entries: WatchingEntryView[],
  localShows: ShowGroup[],
): WatchingShowItem[] {
  const localByAnilistId = new Map<number, ShowGroup>()
  for (const show of localShows) {
    if (!show.bound) {
      continue
    }
    const match = show.key.match(/^anilist:(\d+)$/)
    if (!match) {
      continue
    }
    localByAnilistId.set(Number(match[1]), show)
  }

  return entries.map((entry) => {
    const localShow = localByAnilistId.get(entry.mediaId)
    const maxLocalEpisode = maxLocalEpisodeNumber(localShow)
    const airedLatest = airedLatestEpisode(entry)
    const targetEpisode = Math.max(entry.progress, maxLocalEpisode) + 1
    const newEpisodeNumber =
      airedLatest > 0 &&
      targetEpisode <= airedLatest &&
      !hasLocalEpisode(localShow, targetEpisode)
        ? targetEpisode
        : null

    return {
      mediaId: entry.mediaId,
      key: `anilist:${entry.mediaId}`,
      title: showTitle(entry),
      coverImage: entry.coverImage || localShow?.coverImage || '',
      bannerImage: entry.bannerImage,
      progress: entry.progress,
      totalEpisodes: entry.totalEpisodes,
      mediaStatus: entry.mediaStatus,
      nextAiringEpisode: entry.nextAiringEpisode,
      hasLocalFiles: Boolean(localShow && localShow.episodes.length > 0),
      localShowKey: localShow?.key ?? null,
      maxLocalEpisode,
      newEpisodeNumber,
    }
  })
}

export function watchingPosterCaption(item: WatchingShowItem): {
  text: string
  accent: boolean
} {
  if (item.newEpisodeNumber !== null) {
    return {
      text: `Episode ${item.newEpisodeNumber} available`,
      accent: true,
    }
  }
  if (item.totalEpisodes > 0) {
    const remaining = item.totalEpisodes - item.progress
    if (remaining > 0) {
      return {
        text: `${item.progress} / ${item.totalEpisodes} · ${remaining} left`,
        accent: false,
      }
    }
    return {
      text: `${item.progress} / ${item.totalEpisodes}`,
      accent: false,
    }
  }
  return {
    text: `${item.progress} watched`,
    accent: false,
  }
}

export function watchingPosterSubcaption(
  item: WatchingShowItem,
): string | null {
  if (item.hasLocalFiles) {
    return `${item.maxLocalEpisode} local`
  }
  if (item.newEpisodeNumber !== null) {
    return 'Catch up'
  }
  if (item.mediaStatus === 'RELEASING') {
    return 'Airing'
  }
  if (item.mediaStatus === 'FINISHED') {
    return 'Finished'
  }
  return null
}

export function torrentSearchQuery(
  title: string,
  episodeNumber: number,
): string {
  const paddedEpisode = String(episodeNumber).padStart(2, '0')
  return `${title} ${paddedEpisode}`
}

export function showKeyForEpisodeId(
  localShows: ShowGroup[],
  episodeId: number,
): string | null {
  if (episodeId <= 0) {
    return null
  }
  for (const show of localShows) {
    if (show.episodes.some((episode) => episode.id === episodeId)) {
      return show.key
    }
  }
  return null
}

export function lastWatchedEpisodeIdFromLibrary(
  episodes: EpisodeView[],
): number | null {
  let latestEpisodeId: number | null = null
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
    latestEpisodeId = episode.id
  }

  if (latestEpisodeId !== null) {
    return latestEpisodeId
  }

  let resumeEpisodeId: number | null = null
  let highestResumePosition = 0
  for (const episode of episodes) {
    if (episode.resumePosition <= highestResumePosition) {
      continue
    }
    highestResumePosition = episode.resumePosition
    resumeEpisodeId = episode.id
  }

  return resumeEpisodeId
}

export function pickContinueHeroKey(
  entries: WatchingEntryView[],
  localShows: ShowGroup[],
  playingShowKey: string | null,
  lastPlaybackEpisodeId: number | null = null,
  libraryEpisodes: EpisodeView[] = [],
): string | null {
  if (playingShowKey) {
    return playingShowKey
  }
  const rememberedEpisodeId =
    lastPlaybackEpisodeId ?? lastWatchedEpisodeIdFromLibrary(libraryEpisodes)
  const lastWatchedShowKey = showKeyForEpisodeId(
    localShows,
    rememberedEpisodeId ?? 0,
  )
  if (lastWatchedShowKey) {
    return lastWatchedShowKey
  }
  const items = buildWatchingShowItems(entries, localShows)
  const candidate = items.find((item) => {
    if (!item.hasLocalFiles) {
      return false
    }
    if (item.totalEpisodes > 0 && item.progress >= item.totalEpisodes) {
      return false
    }
    return true
  })
  return candidate?.key ?? null
}
