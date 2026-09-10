import {useEffect, useState} from 'react'
import {GetAnime} from '../../wailsjs/go/main/App'
import {sanitizeAnilistSynopsis} from '../lib/anilistDescription'
import {dayFormatter, timeFormatter} from '../lib/calendar'
import {errorMessage} from '../lib/format'
import type {AnimeView} from '../lib/types'
import {Alert, AlertDescription, AlertTitle} from '@/components/ui/alert'
import {Badge} from '@/components/ui/badge'
import {Skeleton} from '@/components/ui/skeleton'

type Props = {
  anilistId: number
}

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

export function LibraryAnimeDetail({anilistId}: Props) {
  const [anime, setAnime] = useState<AnimeView | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    if (anilistId <= 0) {
      setAnime(null)
      setError('')
      setLoading(false)
      return
    }

    let cancelled = false
    setLoading(true)
    setError('')
    void (async () => {
      try {
        const result = await GetAnime(anilistId)
        if (!cancelled) {
          setAnime(result)
        }
      } catch (err) {
        if (!cancelled) {
          setAnime(null)
          setError(errorMessage(err))
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    })()

    return () => {
      cancelled = true
    }
  }, [anilistId])

  const nextAiringAt = anime?.nextAiringAt ?? 0

  useEffect(() => {
    if (nextAiringAt <= 0) {
      return
    }
    const timer = window.setInterval(() => setNow(Date.now()), 1000)
    return () => window.clearInterval(timer)
  }, [nextAiringAt])

  if (loading) {
    return (
      <section className="mb-6 border border-border bg-bezel p-5">
        <div className="flex flex-col gap-5 sm:flex-row">
          <Skeleton className="mx-auto aspect-2/3 w-64 shrink-0 sm:mx-0" />
          <div className="flex-1 space-y-3">
            <Skeleton className="h-4 w-2/3" />
            <Skeleton className="h-4 w-1/2" />
            <Skeleton className="h-4 w-3/4" />
          </div>
        </div>
      </section>
    )
  }

  if (error) {
    return (
      <Alert variant="destructive" className="mb-6 p-5">
        <AlertTitle>Could not load anime details</AlertTitle>
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    )
  }

  if (!anime) {
    return null
  }

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
  const displayTitle = anime.titleEnglish || anime.titleRomaji
  const romajiTitle = anime.titleRomaji.trim()
  const showRomaji = romajiTitle.length > 0 && romajiTitle !== displayTitle

  return (
    <section className="mb-6 border border-border bg-bezel p-5">
      <div className="mb-5 min-w-0">
        <h3 className="text-xl font-semibold">{displayTitle}</h3>
        {showRomaji && (
          <p className="mt-1 text-base text-muted-foreground">{romajiTitle}</p>
        )}
      </div>
      <div className="flex flex-col gap-5 sm:flex-row">
        {anime.coverImage ? (
          <img
            src={anime.coverImage}
            alt=""
            width={256}
            height={384}
            referrerPolicy="no-referrer"
            className="flex-1 mx-auto aspect-2/3 w-48 shrink-0 object-cover sm:mx-0"
          />
        ) : null}
        <div className="min-w-0 flex-2">
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
          {hasNextAiring && (
            <div className="mt-4">
              <p className="text-sm text-muted-foreground">Next episode</p>
              <p className="font-medium">
                Episode {anime.nextAiringEpisode} · {airingScheduleLabel}
              </p>
              <p>
                {remainingSeconds > 0 ? `Airs in ${countdown}` : 'Airing now'}
              </p>
            </div>
          )}
          {anime.genres.length > 0 && (
            <ul className="mt-4 flex flex-wrap gap-2">
              {anime.genres.map((genre) => (
                <li key={genre}>
                  <Badge variant="outline">{genre}</Badge>
                </li>
              ))}
            </ul>
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
      </div>
    </section>
  )
}
