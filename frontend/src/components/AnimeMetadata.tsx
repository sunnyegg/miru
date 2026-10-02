import {useEffect, useState} from 'react'
import {sanitizeAnilistSynopsis} from '../lib/anilistDescription'
import {dayFormatter, timeFormatter} from '../lib/calendar'
import type {AnimeView} from '../lib/types'
import {Badge} from '@/components/ui/badge'

const formatLabels: Record<string, string> = {
  TV: 'TV',
  TV_SHORT: 'TV Short',
  MOVIE: 'Movie',
  SPECIAL: 'Special',
  OVA: 'OVA',
  ONA: 'ONA',
  MUSIC: 'Music',
}

const statusLabels: Record<string, string> = {
  RELEASING: 'Airing',
  FINISHED: 'Finished',
  NOT_YET_RELEASED: 'Not yet aired',
  CANCELLED: 'Cancelled',
  HIATUS: 'On hiatus',
}

const seasonLabels: Record<string, string> = {
  WINTER: 'Winter',
  SPRING: 'Spring',
  SUMMER: 'Summer',
  FALL: 'Fall',
}

function humanizeEnum(value: string): string {
  return value
    .toLowerCase()
    .split('_')
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ')
}

export function AnimeMetadata({anime}: {anime: AnimeView}) {
  const [now, setNow] = useState(() => Date.now())
  useEffect(() => {
    if (anime.nextAiringAt <= 0) return
    setNow(Date.now())
    const timer = window.setInterval(() => setNow(Date.now()), 1000)
    return () => window.clearInterval(timer)
  }, [anime.nextAiringAt])

  const seasonLabel = [
    seasonLabels[anime.season] ?? humanizeEnum(anime.season),
    anime.seasonYear > 0 ? String(anime.seasonYear) : '',
  ]
    .filter(Boolean)
    .join(' ')

  const rows = [
    {
      label: 'Status',
      value: statusLabels[anime.status] ?? humanizeEnum(anime.status),
    },
    {
      label: 'Format',
      value: formatLabels[anime.format] ?? humanizeEnum(anime.format),
    },
    {
      label: 'Episodes',
      value: anime.totalEpisodes > 0 ? String(anime.totalEpisodes) : '',
    },
    {
      label: 'Duration',
      value: anime.duration > 0 ? `${anime.duration} min` : '',
    },
    {label: 'Season', value: seasonLabel},
    {
      label: 'Score',
      value: anime.averageScore > 0 ? `${anime.averageScore} / 100` : '',
    },
    {
      label: 'Popularity',
      value: anime.popularity > 0 ? anime.popularity.toLocaleString() : '',
    },
    {
      label: 'Favourites',
      value: anime.favourites > 0 ? anime.favourites.toLocaleString() : '',
    },
    {label: 'Source', value: anime.source ? humanizeEnum(anime.source) : ''},
    {label: 'Studios', value: anime.studios.join(', ')},
    {label: 'Japanese', value: anime.titleNative},
  ].filter((row) => row.value)

  const airingAt = anime.nextAiringAt
  const airingDate = new Date(airingAt * 1000)
  const remainingSeconds = Math.max(
    0,
    Math.round((airingAt * 1000 - now) / 1000),
  )
  const days = Math.floor(remainingSeconds / 86400)
  const hours = Math.floor((remainingSeconds % 86400) / 3600)
  const minutes = Math.floor((remainingSeconds % 3600) / 60)
  const seconds = remainingSeconds % 60
  const countdown = [
    days > 0 ? `${days}d` : '',
    hours > 0 ? `${hours}h` : '',
    minutes > 0 ? `${minutes}m` : '',
    days === 0 ? `${seconds}s` : '',
  ]
    .filter(Boolean)
    .join(' ')
  const hasNextAiring = airingAt > 0 && anime.nextAiringEpisode > 0
  const airingScheduleLabel = `${dayFormatter.format(airingDate)} · ${timeFormatter.format(airingDate)}`
  return (
    <div>
      {rows.length > 0 && (
        <dl className="grid grid-cols-2 gap-x-6 gap-y-3 sm:grid-cols-3">
          {rows.map((row) => (
            <div key={row.label} className="min-w-0">
              <dt className="text-sm text-muted-foreground">{row.label}</dt>
              <dd className="font-medium break-words">{row.value}</dd>
            </div>
          ))}
        </dl>
      )}
      {anime.genres.length > 0 && (
        <div className="mt-4">
          <p className="text-sm text-muted-foreground">Genre</p>
          <ul className="mt-1 flex flex-wrap gap-2">
            {anime.genres.map((genre) => (
              <li key={genre}>
                <Badge variant="outline">{genre}</Badge>
              </li>
            ))}
          </ul>
        </div>
      )}
      {hasNextAiring && (
        <div className="mt-4">
          <p className="text-sm text-muted-foreground">Next episode</p>
          <p className="font-medium">
            Episode {anime.nextAiringEpisode} · {airingScheduleLabel}
          </p>
          <p>{remainingSeconds > 0 ? `Airs in ${countdown}` : 'Airing now'}</p>
        </div>
      )}
      {anime.synopsis.trim() && (
        <div className="mt-4">
          <p className="text-sm text-muted-foreground">Synopsis</p>
          <div
            className="mt-1 max-h-72 overflow-y-auto text-foreground/90 [&_a]:text-accent [&_a]:underline"
            dangerouslySetInnerHTML={{
              __html: sanitizeAnilistSynopsis(anime.synopsis),
            }}
          />
        </div>
      )}
    </div>
  )
}
