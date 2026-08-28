import type {ShowGroup} from '../lib/groupEpisodes'
import {LibraryPosterGrid} from './LibraryPosterGrid'

type Props = {
  loading: boolean
  loadError: string
  shows: ShowGroup[]
  highlightedKey: string | null
  onSelectShow: (key: string) => void
  onRetry: () => void
  suppressEmptyState?: boolean
}

export function LibraryUnlistedSection({
  loading,
  loadError,
  shows,
  highlightedKey,
  onSelectShow,
  onRetry,
  suppressEmptyState = false,
}: Props) {
  if (!loading && shows.length === 0 && suppressEmptyState) {
    return null
  }

  return (
    <section className="shrink-0">
      <h3 className="mb-3 text-sm font-medium text-muted-foreground">Unlisted</h3>
      <LibraryPosterGrid
        loading={loading}
        loadError={loadError}
        shows={shows}
        highlightedKey={highlightedKey}
        onSelectShow={onSelectShow}
        onRetry={onRetry}
        suppressEmptyState={suppressEmptyState}
      />
    </section>
  )
}
