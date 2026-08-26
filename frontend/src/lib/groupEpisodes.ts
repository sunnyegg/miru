import type {EpisodeView} from './types'

export type ShowGroup = {
  key: string
  title: string
  coverImage: string
  bound: boolean
  unlinkedCount: number
  episodes: EpisodeView[]
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
  }

  const shows = [...groups.values()]
  for (const show of shows) {
    show.episodes.sort((a, b) => a.episodeNumber - b.episodeNumber)
  }
  return shows
}
