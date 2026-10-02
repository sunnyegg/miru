import {useEffect, useState} from 'react'
import {GetAnime} from '../../wailsjs/go/main/App'
import {AnimeMetadata} from './AnimeMetadata'
import {errorMessage} from '../lib/format'
import type {AnimeView} from '../lib/types'
import {Alert, AlertDescription, AlertTitle} from '@/components/ui/alert'
import {Skeleton} from '@/components/ui/skeleton'

type Props = {
  anilistId: number
}

export function LibraryAnimeDetail({anilistId}: Props) {
  const [anime, setAnime] = useState<AnimeView | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

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
    async function loadAnimeDetails() {
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
    }
    void loadAnimeDetails()

    return () => {
      cancelled = true
    }
  }, [anilistId])

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
          <AnimeMetadata anime={anime} />
        </div>
      </div>
    </section>
  )
}
