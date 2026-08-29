import type {ShowGroup} from '../lib/groupEpisodes'
import {LibraryPosterCard} from './LibraryPosterCard'
import {Alert} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'
import {Skeleton} from '@/components/ui/skeleton'

function posterCaption(show: ShowGroup): string {
  if (!show.bound) {
    const count = show.episodes.length
    return `${count} episode${count === 1 ? '' : 's'} · not linked to AniList`
  }
  if (show.totalEpisodes > 0) {
    const remaining = show.totalEpisodes - show.progress
    if (remaining > 0) {
      return `${show.progress} / ${show.totalEpisodes} · ${remaining} left`
    }
    return `${show.progress} / ${show.totalEpisodes}`
  }
  return `${show.progress} watched`
}

function posterSubcaption(show: ShowGroup): string | null {
  if (
    show.bound &&
    show.totalEpisodes > 0 &&
    show.progress < show.totalEpisodes
  ) {
    return `Next: Ep ${show.progress + 1}`
  }
  if (!show.bound) {
    return 'Click to match AniList'
  }
  return null
}

type Props = {
  loading: boolean
  loadError: string
  shows: ShowGroup[]
  highlightedKey: string | null
  onSelectShow: (key: string) => void
  onRetry: () => void
  suppressEmptyState?: boolean
}

export function LibraryPosterGrid({
  loading,
  loadError,
  shows,
  highlightedKey,
  onSelectShow,
  onRetry,
  suppressEmptyState = false,
}: Props) {
  if (loading) {
    return (
      <ul
        className="grid grid-cols-[repeat(auto-fill,minmax(min(12rem,100%),1fr))] gap-5 p-1"
        aria-busy="true"
        aria-label="Loading library"
      >
        {Array.from({length: 8}, (_, index) => (
          <li key={index}>
            <Skeleton className="aspect-[2/3] w-full animate-pulse" />
          </li>
        ))}
      </ul>
    )
  }

  if (loadError && shows.length === 0) {
    return (
      <Alert
        variant="destructive"
        className="flex min-h-48 flex-wrap items-end justify-between gap-4 p-4"
      >
        <div className="min-w-0">
          <h3 className="font-medium text-foreground">
            Library could not be loaded
          </h3>
          <p className="mt-1 wrap-break-word text-sm text-destructive">
            {loadError}
          </p>
        </div>
        <Button type="button" variant="secondary" onClick={onRetry}>
          Try again
        </Button>
      </Alert>
    )
  }

  if (shows.length === 0) {
    if (suppressEmptyState) {
      return null
    }
    return (
      <div className="flex min-h-32 items-end">
        <p className="max-w-md text-sm text-muted-foreground">
          No local shows yet. Import a file, or finish a torrent download and it
          will land here.
        </p>
      </div>
    )
  }

  return (
    <ul className="grid grid-cols-[repeat(auto-fill,minmax(min(12rem,100%),1fr))] gap-5 p-1">
      {shows.map((show) => (
        <li key={show.key}>
          <LibraryPosterCard
            title={show.title}
            coverImage={show.coverImage}
            caption={posterCaption(show)}
            subcaption={posterSubcaption(show)}
            active={show.key === highlightedKey}
            size="grid"
            onClick={() => onSelectShow(show.key)}
          />
        </li>
      ))}
    </ul>
  )
}
