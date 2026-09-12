import {
  watchingPosterCaption,
  watchingPosterSubcaption,
  type WatchingShowItem,
} from '../lib/libraryWatching'
import {compareByNextAiring} from '../lib/groupEpisodes'
import {LibraryPosterCard} from './LibraryPosterCard'
import {LibraryPosterCarousel} from './LibraryPosterCarousel'
import {Skeleton} from '@/components/ui/skeleton'

type Props = {
  title: string
  emptyMessage: string
  items: WatchingShowItem[]
  loading: boolean
  highlightedKey: string | null
  excludeHeroKey?: string | null
  onOpenShow: (showKey: string) => void
}

export function LibraryWatchingSection({
  title,
  emptyMessage,
  items: sourceItems,
  loading,
  highlightedKey,
  excludeHeroKey = null,
  onOpenShow,
}: Props) {
  if (!loading && sourceItems.length === 0) {
    return (
      <section className="mb-8 shrink-0">
        <div className="mb-3 flex items-baseline gap-2">
          <h3 className="text-sm font-medium text-foreground">{title}</h3>
        </div>
        <p className="text-sm text-muted-foreground">{emptyMessage}</p>
      </section>
    )
  }

  const items = (
    excludeHeroKey
      ? sourceItems.filter((item) => item.key !== excludeHeroKey)
      : [...sourceItems]
  ).sort((left, right) => compareByNextAiring(left, right, Date.now()))

  return (
    <section className="mb-8 shrink-0">
      <div className="mb-3 flex items-baseline gap-2">
        <h3 className="text-sm font-medium text-foreground">{title}</h3>
        {!loading && sourceItems.length > 0 && (
          <span className="text-xs text-muted-foreground">
            {sourceItems.length}
          </span>
        )}
      </div>
      {loading ? (
        <LibraryPosterCarousel ariaLabel={`Loading ${title} list`} ariaBusy>
          {Array.from({length: 4}, (_, index) => (
            <li key={index} className="w-44 shrink-0 sm:w-48">
              <Skeleton className="aspect-[2/3] w-full animate-pulse" />
            </li>
          ))}
        </LibraryPosterCarousel>
      ) : items.length === 0 ? null : (
        <LibraryPosterCarousel ariaLabel={`${title} shelf`}>
          {items.map((item) => {
            const caption = watchingPosterCaption(item)
            const subcaption = watchingPosterSubcaption(item)

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
                  onClick={() => onOpenShow(item.key)}
                />
              </li>
            )
          })}
        </LibraryPosterCarousel>
      )}
    </section>
  )
}
