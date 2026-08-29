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
      <div className="mb-3 flex items-baseline gap-2">
        <h3 className="text-sm font-medium text-foreground">Local library</h3>
        {!loading && shows.length > 0 && (
          <span className="text-xs text-muted-foreground">{shows.length}</span>
        )}
      </div>
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
