import type {EpisodeView} from './types'

export type ShowGroup = {
  key: string
  title: string
  coverImage: string
  bound: boolean
  unlinkedCount: number
  progress: number
  totalEpisodes: number
  mediaStatus: string
  nextAiringEpisode: number
  episodes: EpisodeView[]
}

export type EpisodeSlotKind = 'available' | 'missing' | 'upcoming'

export type EpisodeSlot = {
  number: number
  file: EpisodeView | null
  kind: EpisodeSlotKind
}

export function visibleLibraryEpisodes(episodes: EpisodeView[]): EpisodeView[] {
  const boundTitles = new Set(
    episodes.filter((episode) => episode.bound).map((episode) => episode.displayTitle),
  )
  const seenUnboundTitles = new Set<string>()
  return episodes.filter((episode) => {
    if (episode.bound) {
      return true
    }
    if (boundTitles.has(episode.displayTitle) || seenUnboundTitles.has(episode.displayTitle)) {
      return false
    }
    seenUnboundTitles.add(episode.displayTitle)
    return true
  })
}

export function groupEpisodes(episodes: EpisodeView[]): ShowGroup[] {
  const groups = new Map<string, ShowGroup>()

  for (const episode of episodes) {
    const key = episode.anilistId > 0
      ? `anilist:${episode.anilistId}`
      : `file:${episode.animeTitle || episode.displayTitle}`
    const existing = groups.get(key)
    if (!existing) {
      groups.set(key, {
        key,
        title: episode.animeTitle || episode.displayTitle,
        coverImage: episode.coverImage,
        bound: episode.bound,
        unlinkedCount: episode.bound ? 0 : 1,
        progress: episode.progress,
        totalEpisodes: episode.totalEpisodes,
        mediaStatus: episode.mediaStatus,
        nextAiringEpisode: episode.nextAiringEpisode,
        episodes: [episode],
      })
      continue
    }

    existing.episodes.push(episode)
    if (!episode.bound) {
      existing.unlinkedCount += 1
    }
    if (episode.bound) {
      existing.bound = true
    }
    if (!existing.coverImage && episode.coverImage) {
      existing.coverImage = episode.coverImage
    }
    if (episode.animeTitle) {
      existing.title = episode.animeTitle
    }
    if (episode.mediaStatus) {
      existing.mediaStatus = episode.mediaStatus
    }
    if (episode.nextAiringEpisode > 0) {
      existing.nextAiringEpisode = episode.nextAiringEpisode
    }
    if (episode.totalEpisodes > 0) {
      existing.totalEpisodes = episode.totalEpisodes
    }
  }

  const shows = [...groups.values()]
  for (const show of shows) {
    show.episodes.sort((a, b) => a.episodeNumber - b.episodeNumber)
  }
  return shows
}

export function episodeSlots(show: ShowGroup): EpisodeSlot[] {
  if (!show.bound) {
    return show.episodes.map((episode) => ({
      number: episode.episodeNumber,
      file: episode,
      kind: 'available' as const,
    }))
  }

  const filesByNumber = new Map<number, EpisodeView>()
  let maxLocalNumber = 0

  for (const episode of show.episodes) {
    if (episode.episodeNumber > 0) {
      filesByNumber.set(episode.episodeNumber, episode)
      if (episode.episodeNumber > maxLocalNumber) {
        maxLocalNumber = episode.episodeNumber
      }
    }
  }

  let slotCount = show.totalEpisodes
  if (slotCount <= 0) {
    slotCount = Math.max(show.nextAiringEpisode, maxLocalNumber)
  }

  const slots: EpisodeSlot[] = []
  const hasUpcoming = show.nextAiringEpisode > 0 && show.mediaStatus !== 'FINISHED'

  for (let number = 1; number <= slotCount; number++) {
    const file = filesByNumber.get(number) ?? null
    if (file) {
      slots.push({number, file, kind: 'available'})
      continue
    }
    if (hasUpcoming && number >= show.nextAiringEpisode) {
      slots.push({number, file: null, kind: 'upcoming'})
      continue
    }
    slots.push({number, file: null, kind: 'missing'})
  }

  for (const episode of show.episodes) {
    if (episode.episodeNumber <= 0 || episode.episodeNumber > slotCount) {
      slots.push({
        number: episode.episodeNumber,
        file: episode,
        kind: 'available',
      })
    }
  }

  slots.sort((left, right) => {
    if (left.number <= 0 && right.number <= 0) {
      return 0
    }
    if (left.number <= 0) {
      return 1
    }
    if (right.number <= 0) {
      return -1
    }
    return left.number - right.number
  })

  return slots
}
