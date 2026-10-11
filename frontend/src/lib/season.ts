const seasonLabels: Record<string, string> = {
  WINTER: 'Winter',
  SPRING: 'Spring',
  SUMMER: 'Summer',
  FALL: 'Fall',
}

const seasonOrder: Record<string, number> = {
  WINTER: 1,
  SPRING: 2,
  SUMMER: 3,
  FALL: 4,
}

export const mediaSeasons = ['WINTER', 'SPRING', 'SUMMER', 'FALL'] as const

export type MediaSeason = (typeof mediaSeasons)[number]

export type SeasonGroup<T> = {
  key: string
  label: string
  items: T[]
}

export function seasonDisplayName(season: string): string {
  const normalized = season.toUpperCase()
  if (seasonLabels[normalized]) {
    return seasonLabels[normalized]
  }
  if (!season) {
    return ''
  }
  return season
    .toLowerCase()
    .split('_')
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ')
}

export function seasonLabel(season: string, seasonYear: number): string {
  const name = seasonDisplayName(season)
  if (name && seasonYear > 0) {
    return `${name} ${seasonYear}`
  }
  if (name) {
    return name
  }
  if (seasonYear > 0) {
    return String(seasonYear)
  }
  return 'Unknown'
}

export function seasonSortKey(season: string, seasonYear: number): number {
  if (!season || seasonYear <= 0) {
    return Number.NEGATIVE_INFINITY
  }
  const order = seasonOrder[season.toUpperCase()] ?? 0
  return seasonYear * 10 + order
}

export function seasonGroupKey(season: string, seasonYear: number): string {
  if (!season || seasonYear <= 0) {
    return 'unknown'
  }
  return `${seasonYear}:${season.toUpperCase()}`
}

export function groupBySeason<T>(
  items: T[],
  getSeason: (item: T) => {season: string; seasonYear: number},
): SeasonGroup<T>[] {
  const groups = new Map<string, SeasonGroup<T>>()
  for (const item of items) {
    const {season, seasonYear} = getSeason(item)
    const key = seasonGroupKey(season, seasonYear)
    const existing = groups.get(key)
    if (existing) {
      existing.items.push(item)
      continue
    }
    groups.set(key, {
      key,
      label: seasonLabel(season, seasonYear),
      items: [item],
    })
  }

  return [...groups.values()].sort((left, right) => {
    const leftSeason = getSeason(left.items[0])
    const rightSeason = getSeason(right.items[0])
    return (
      seasonSortKey(rightSeason.season, rightSeason.seasonYear) -
      seasonSortKey(leftSeason.season, leftSeason.seasonYear)
    )
  })
}

export function currentAniListSeason(date: Date = new Date()): {
  season: MediaSeason
  year: number
} {
  const month = date.getMonth() + 1
  const year = date.getFullYear()
  if (month <= 3) {
    return {season: 'WINTER', year}
  }
  if (month <= 6) {
    return {season: 'SPRING', year}
  }
  if (month <= 9) {
    return {season: 'SUMMER', year}
  }
  return {season: 'FALL', year}
}
