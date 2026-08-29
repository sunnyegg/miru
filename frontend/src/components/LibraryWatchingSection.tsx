import {
  buildWatchingShowItems,
  torrentSearchQuery,
  watchingPosterCaption,
  watchingPosterSubcaption,
} from '../lib/libraryWatching'
import type {ShowGroup} from '../lib/groupEpisodes'
import type {WatchingEntryView} from '../lib/types'
import {LibraryPosterCard} from './LibraryPosterCard'
import {LibraryPosterCarousel} from './LibraryPosterCarousel'
import {Skeleton} from '@/components/ui/skeleton'

type Props = {
  entries: WatchingEntryView[]
  localShows: ShowGroup[]
  loading: boolean
  highlightedKey: string | null
  excludeHeroKey?: string | null
  onOpenShow: (localShowKey: string) => void
  onFindTorrent: (query: string) => void
}

export function LibraryWatchingSection({
  entries,
  localShows,
  loading,
  highlightedKey,
  excludeHeroKey = null,
  onOpenShow,
  onFindTorrent,
}: Props) {
  if (!loading && entries.length === 0) {
    return (
      <section className="mb-8 shrink-0">
        <div className="mb-3 flex items-baseline gap-2">
          <h3 className="text-sm font-medium text-foreground">Watching</h3>
        </div>
        <p className="text-sm text-muted-foreground">
          Nothing on your AniList Watching list yet. Add titles from the
          Watching tab to track them here.
        </p>
      </section>
    )
  }

  const allItems = buildWatchingShowItems(entries, localShows)
  const items = excludeHeroKey
    ? allItems.filter((item) => item.key !== excludeHeroKey)
    : allItems

  return (
    <section className="mb-8 shrink-0">
      <div className="mb-3 flex items-baseline gap-2">
        <h3 className="text-sm font-medium text-foreground">Watching</h3>
        {!loading && allItems.length > 0 && (
          <span className="text-xs text-muted-foreground">
            {allItems.length}
          </span>
        )}
      </div>
      {loading ? (
        <LibraryPosterCarousel ariaLabel="Loading Watching list" ariaBusy>
          {Array.from({length: 4}, (_, index) => (
            <li key={index} className="w-44 shrink-0 sm:w-48">
              <Skeleton className="aspect-[2/3] w-full animate-pulse" />
            </li>
          ))}
        </LibraryPosterCarousel>
      ) : items.length === 0 ? null : (
        <LibraryPosterCarousel ariaLabel="Watching shelf">
          {items.map((item) => {
            const caption = watchingPosterCaption(item)
            const subcaption = watchingPosterSubcaption(item)

            function handleClick() {
              if (item.hasLocalFiles && item.localShowKey) {
                onOpenShow(item.localShowKey)
                return
              }
              const episodeNumber = item.newEpisodeNumber ?? item.progress + 1
              onFindTorrent(torrentSearchQuery(item.title, episodeNumber))
            }

            return (
              <li key={item.key}>
                <LibraryPosterCard
                  title={item.title}
                  coverImage={item.coverImage}
                  caption={caption.text}
                  subcaption={subcaption}
                  accentCaption={caption.accent}
                  active={item.key === highlightedKey}
                  size="shelf"
                  onClick={handleClick}
                />
              </li>
            )
          })}
        </LibraryPosterCarousel>
      )}
    </section>
  )
}
