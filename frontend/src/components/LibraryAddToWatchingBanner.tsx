import type {ShowGroup} from '../lib/groupEpisodes'
import {Button} from '@/components/ui/button'
import {Card} from '@/components/ui/card'

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
    <Card className="mb-3 border border-border" role="region" aria-labelledby="add-watching-title">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h3 id="add-watching-title" className="text-base font-medium">Add to Watching</h3>
          {mediaId ? (
            <p className="mt-1 text-sm text-muted-foreground">
              {show.title} is in your library but not on your AniList Watching list. Add it to
              track new episodes and catch up from here.
            </p>
          ) : (
            <p className="mt-1 text-sm text-muted-foreground">
              Match {show.title} to an AniList title first, then add it to your Watching list.
            </p>
          )}
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
    </Card>
  )
}
