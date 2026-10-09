import type {ShowGroup} from './groupEpisodes'
import {nextAiringLabel} from './calendar'
import type {EpisodeView, WatchingEntryView} from './types'

export type WatchingShowItem = {
  mediaId: number
  key: string
  title: string
  titleRomaji: string
  titleEnglish: string
  titleNative: string
  coverImage: string
  bannerImage: string
  progress: number
  totalEpisodes: number
  mediaStatus: string
  nextAiringEpisode: number
  nextAiringAt: number
  hasLocalFiles: boolean
  maxLocalEpisode: number
  newEpisodeNumber: number | null
}

export type TorrentSearchTitles = {
  romaji: string
  english: string
  native: string
  fallback: string
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
      titleRomaji: entry.titleRomaji,
      titleEnglish: entry.titleEnglish,
      titleNative: entry.titleNative || '',
      coverImage: entry.coverImage || localShow?.coverImage || '',
      bannerImage: entry.bannerImage,
      progress: entry.progress,
      totalEpisodes: entry.totalEpisodes,
      mediaStatus: entry.mediaStatus,
      nextAiringEpisode: entry.nextAiringEpisode,
      nextAiringAt: entry.nextAiringAt,
      hasLocalFiles: Boolean(localShow && localShow.episodes.length > 0),
      maxLocalEpisode,
      newEpisodeNumber,
    }
  })
}

export function isWatchingItemAvailable(item: WatchingShowItem): boolean {
  return item.hasLocalFiles || item.newEpisodeNumber !== null
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
  const airing = nextAiringLabel(item.nextAiringEpisode, item.nextAiringAt)
  if (item.totalEpisodes > 0) {
    const base = `${item.progress} / ${item.totalEpisodes}`
    if (airing) {
      return {
        text: `${base} · ${airing}`,
        accent: false,
      }
    }
    const remaining = item.totalEpisodes - item.progress
    if (remaining > 0) {
      return {
        text: `${base} · ${remaining} left`,
        accent: false,
      }
    }
    return {
      text: base,
      accent: false,
    }
  }
  const base = `${item.progress} watched`
  return {
    text: airing ? `${base} · ${airing}` : base,
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

export function torrentSearchTitlesFromEntry(
  entry: {
    titleRomaji: string
    titleEnglish: string
    titleNative?: string
  } | null,
  fallback: string,
): TorrentSearchTitles {
  return {
    romaji: entry?.titleRomaji ?? '',
    english: entry?.titleEnglish ?? '',
    native: entry?.titleNative ?? '',
    fallback,
  }
}

export function torrentSearchTitlesFromWatchingItem(
  item: WatchingShowItem,
): TorrentSearchTitles {
  return {
    romaji: item.titleRomaji,
    english: item.titleEnglish,
    native: item.titleNative,
    fallback: item.title,
  }
}

export function pickTorrentSearchTitle(
  titles: TorrentSearchTitles,
  language: string,
): string {
  const preferred =
    language === 'english'
      ? titles.english
      : language === 'native'
        ? titles.native
        : titles.romaji
  return (
    preferred.trim() ||
    titles.romaji.trim() ||
    titles.english.trim() ||
    titles.native.trim() ||
    titles.fallback.trim()
  )
}

export function withTorrentSearchPrefix(query: string, prefix: string): string {
  const trimmedQuery = query.trim()
  const trimmedPrefix = prefix.trim()
  if (!trimmedQuery) {
    return ''
  }
  if (!trimmedPrefix) {
    return trimmedQuery
  }
  if (trimmedQuery.startsWith(trimmedPrefix)) {
    return trimmedQuery
  }
  return `${trimmedPrefix} ${trimmedQuery}`
}

export function torrentSearchQuery(
  title: string,
  episodeNumber: number,
): string {
  const paddedEpisode = String(episodeNumber).padStart(2, '0')
  return `${title} ${paddedEpisode}`
}

export function buildEpisodeShowKeyMap(
  localShows: ShowGroup[],
): Map<number, string> {
  const showKeyByEpisodeId = new Map<number, string>()
  for (const show of localShows) {
    for (const episode of show.episodes) {
      showKeyByEpisodeId.set(episode.id, show.key)
    }
  }
  return showKeyByEpisodeId
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
  episodeShowKeys: ReadonlyMap<number, string> = new Map(),
  completedShowKeys: ReadonlySet<string> = new Set(),
): string | null {
  if (playingShowKey) {
    return playingShowKey
  }
  const rememberedEpisodeId =
    lastPlaybackEpisodeId ?? lastWatchedEpisodeIdFromLibrary(libraryEpisodes)
  const lastWatchedShowKey =
    episodeShowKeys.get(rememberedEpisodeId ?? 0) ?? null
  if (lastWatchedShowKey) {
    return completedShowKeys.has(lastWatchedShowKey) ? null : lastWatchedShowKey
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
