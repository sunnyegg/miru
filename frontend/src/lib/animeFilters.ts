import type {AnimeAdvancedSearchInput} from './types'

// AnimeAdvancedFilters is the part of the AniList advanced search request the
// sheet edits. The page number is added by the store per request.
export type AnimeAdvancedFilters = Omit<AnimeAdvancedSearchInput, 'page'>

export const defaultAdvancedFilters: AnimeAdvancedFilters = {
  search: '',
  genres: [],
  excludeGenres: [],
  year: 0,
  season: '',
  formats: [],
  statuses: [],
  minScore: 0,
  sort: 'POPULARITY_DESC',
}

export const formatLabels: Record<string, string> = {
  TV: 'TV',
  TV_SHORT: 'TV short',
  MOVIE: 'Movie',
  SPECIAL: 'Special',
  OVA: 'OVA',
  ONA: 'ONA',
  MUSIC: 'Music',
}

export const statusLabels: Record<string, string> = {
  FINISHED: 'Finished',
  RELEASING: 'Releasing',
  NOT_YET_RELEASED: 'Not yet released',
  CANCELLED: 'Cancelled',
  HIATUS: 'Hiatus',
}

export const sortLabels: Record<string, string> = {
  POPULARITY_DESC: 'Popularity',
  SCORE_DESC: 'Score',
  TRENDING_DESC: 'Trending',
  START_DATE_DESC: 'Newest',
  TITLE_ROMAJI: 'Title A-Z',
}

// advancedFilterCount counts the filters that are set, so the sheet can tell an
// empty request from a real one.
export function advancedFilterCount(filters: AnimeAdvancedFilters): number {
  let count = filters.search.trim() === '' ? 0 : 1
  count += filters.genres.length
  count += filters.excludeGenres.length
  count += filters.year > 0 ? 1 : 0
  count += filters.season === '' ? 0 : 1
  count += filters.formats.length
  count += filters.statuses.length
  count += filters.minScore > 0 ? 1 : 0
  return count
}

// advancedFiltersValid rejects the two shapes AniList would refuse: no filter
// at all, and a season without a year.
export function advancedFiltersValid(filters: AnimeAdvancedFilters): boolean {
  return (
    advancedFilterCount(filters) > 0 &&
    !(filters.season !== '' && filters.year === 0)
  )
}
