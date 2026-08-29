import type {ShowGroup} from '../lib/groupEpisodes'
import {Button} from '@/components/ui/button'

type Props = {
  show: ShowGroup
  saving: boolean
  onAddToWatching: () => void
  onMatchAnilist: () => void
}

function anilistMediaId(show: ShowGroup): number | null {
  if (!show.bound) {
    return null
  }
  const match = show.key.match(/^anilist:(\d+)$/)
  if (!match) {
    return null
  }
  return Number(match[1])
}

export function LibraryAddToWatchingBanner({
  show,
  saving,
  onAddToWatching,
  onMatchAnilist,
}: Props) {
  const mediaId = anilistMediaId(show)

  return (
    <div
      className="flex flex-wrap items-center justify-between gap-3 border border-border/60 bg-card/80 px-4 py-3 backdrop-blur-sm"
      role="region"
      aria-labelledby="add-watching-title"
    >
      <div className="min-w-0">
        <h3 id="add-watching-title" className="text-sm font-medium">Add to Watching</h3>
        <p className="mt-0.5 text-xs text-muted-foreground">
          {mediaId
            ? 'Track new episodes and catch up from your library.'
            : 'Match to AniList first, then add to your Watching list.'}
        </p>
      </div>
      {mediaId ? (
        <Button type="button" onClick={onAddToWatching} disabled={saving}>
          {saving ? 'Saving…' : 'Watch'}
        </Button>
      ) : (
        <Button type="button" variant="secondary" onClick={onMatchAnilist}>
          Match AniList
        </Button>
      )}
    </div>
  )
}
