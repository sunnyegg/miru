import type {ShowGroup} from '../lib/groupEpisodes'
import {Alert} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'
import {Skeleton} from '@/components/ui/skeleton'

function posterCaption(show: ShowGroup): string {
  if (!show.bound) {
    const count = show.episodes.length
    return `${count} episode${count === 1 ? '' : 's'} · not linked`
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

type Props = {
  loading: boolean
  loadError: string
  shows: ShowGroup[]
  highlightedKey: string | null
  onSelectShow: (key: string) => void
  onRetry: () => void
}

export function LibraryPosterGrid({
  loading,
  loadError,
  shows,
  highlightedKey,
  onSelectShow,
  onRetry,
}: Props) {
  if (loading) {
    return (
      <ul className="grid grid-cols-[repeat(auto-fill,minmax(min(8.5rem,100%),1fr))] gap-3 p-1" aria-busy="true" aria-label="Loading library">
        {Array.from({length: 8}, (_, index) => (
          <li key={index}>
            <Skeleton className="aspect-square w-full" />
          </li>
        ))}
      </ul>
    )
  }

  if (loadError && shows.length === 0) {
    return (
      <Alert variant="destructive" className="flex min-h-48 flex-wrap items-end justify-between gap-4 p-4">
        <div className="min-w-0">
          <h3 className="font-medium text-foreground">Library could not be loaded</h3>
          <p className="mt-1 wrap-break-word text-sm text-destructive">{loadError}</p>
        </div>
        <Button type="button" variant="secondary" onClick={onRetry}>
          Try again
        </Button>
      </Alert>
    )
  }

  if (shows.length === 0) {
    return (
      <div className="flex h-full min-h-48 items-end">
        <p className="max-w-md text-sm text-muted-foreground">
          No local shows yet. Import a file, or finish a torrent download and it will land here.
        </p>
      </div>
    )
  }

  return (
    <ul className="grid grid-cols-[repeat(auto-fill,minmax(min(8.5rem,100%),1fr))] gap-3 p-1">
      {shows.map((show) => {
        const active = show.key === highlightedKey
        return (
          <li key={show.key}>
            <button
              type="button"
              onClick={() => onSelectShow(show.key)}
              aria-pressed={active}
              className={`group flex w-full cursor-pointer flex-col text-left ${
                active ? 'outline-2 outline-offset-2 outline-accent' : ''
              }`}
            >
              {show.coverImage ? (
                <img
                  src={show.coverImage}
                  alt=""
                  referrerPolicy="no-referrer"
                  className="aspect-square w-full bg-muted object-cover"
                />
              ) : (
                <span className="flex aspect-square w-full items-end bg-muted p-2 text-xs text-muted-foreground">
                  {show.title}
                </span>
              )}
              <span className="mt-2 truncate text-sm font-medium">{show.title}</span>
              <span className="truncate text-xs text-muted-foreground">
                {posterCaption(show)}
              </span>
            </button>
          </li>
        )
      })}
    </ul>
  )
}
