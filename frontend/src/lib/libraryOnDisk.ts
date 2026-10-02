import type {ShowGroup} from './groupEpisodes'
import type {WatchingEntryView} from './types'

export const onDiskListStatuses = [
  'COMPLETED',
  'PLANNING',
  'PAUSED',
  'DROPPED',
  'REPEATING',
] as const

export type OnDiskFilter =
  | 'ALL'
  | (typeof onDiskListStatuses)[number]
  | 'NOT_ON_LIST'
  | 'NEEDS_MATCHING'

export const onDiskFilters: {value: OnDiskFilter; label: string}[] = [
  {value: 'ALL', label: 'All'},
  {value: 'COMPLETED', label: 'Completed'},
  {value: 'PLANNING', label: 'Planning'},
  {value: 'PAUSED', label: 'Paused'},
  {value: 'DROPPED', label: 'Dropped'},
  {value: 'REPEATING', label: 'Repeating'},
  {value: 'NOT_ON_LIST', label: 'Not on list'},
  {value: 'NEEDS_MATCHING', label: 'Needs matching'},
]

export function bucketOnDiskShows(
  shows: ShowGroup[],
  entries: WatchingEntryView[],
): Record<OnDiskFilter, ShowGroup[]> {
  const buckets = Object.fromEntries(
    onDiskFilters.map(({value}) => [value, [] as ShowGroup[]]),
  ) as Record<OnDiskFilter, ShowGroup[]>
  buckets.ALL = shows

  const statusByMediaId = new Map(
    entries.map((entry) => [entry.mediaId, entry.listStatus]),
  )

  for (const show of shows) {
    let filter: OnDiskFilter = 'NOT_ON_LIST'
    if (!show.bound) {
      filter = 'NEEDS_MATCHING'
    } else {
      const mediaId = show.episodes.find(
        (episode) => episode.anilistId > 0,
      )?.anilistId
      const status = mediaId ? statusByMediaId.get(mediaId) : undefined
      if (
        status &&
        (onDiskListStatuses as readonly string[]).includes(status)
      ) {
        filter = status as OnDiskFilter
      }
    }
    buckets[filter].push(show)
  }

  return buckets
}
