import {
  buildWatchingShowItems,
  torrentSearchQuery,
  watchingPosterCaption,
  watchingPosterSubcaption,
  type WatchingShowItem,
} from '../lib/libraryWatching'
import type {ShowGroup} from '../lib/groupEpisodes'
import type {WatchingEntryView} from '../lib/types'
import {Skeleton} from '@/components/ui/skeleton'

type Props = {
  entries: WatchingEntryView[]
  localShows: ShowGroup[]
  loading: boolean
  highlightedKey: string | null
  onOpenShow: (localShowKey: string) => void
  onFindTorrent: (query: string) => void
}

function WatchingPoster({
  item,
  active,
  onOpenShow,
  onFindTorrent,
}: {
  item: WatchingShowItem
  active: boolean
  onOpenShow: (localShowKey: string) => void
  onFindTorrent: (query: string) => void
}) {
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
    <button
      type="button"
      onClick={handleClick}
      aria-pressed={active}
      className={`group flex w-full cursor-pointer flex-col text-left transition-colors duration-200 motion-reduce:transition-none ${
        active ? 'outline-2 outline-offset-2 outline-accent' : ''
      }`}
    >
      {item.coverImage ? (
        <img
          src={item.coverImage}
          alt=""
          referrerPolicy="no-referrer"
          className="aspect-square w-full bg-muted object-cover"
        />
      ) : (
        <span className="flex aspect-square w-full items-end bg-muted p-2 text-xs text-muted-foreground">
          {item.title}
        </span>
      )}
      <span className="mt-2 truncate text-sm font-medium">{item.title}</span>
      <span
        className={`truncate text-xs ${caption.accent ? 'text-accent' : 'text-muted-foreground'}`}
      >
        {caption.text}
      </span>
      {subcaption && (
        <span className="truncate text-xs text-muted-foreground opacity-0 transition-opacity duration-200 group-hover:opacity-100 motion-reduce:opacity-100 motion-reduce:transition-none">
          {subcaption}
        </span>
      )}
    </button>
  )
}

export function LibraryWatchingSection({
  entries,
  localShows,
  loading,
  highlightedKey,
  onOpenShow,
  onFindTorrent,
}: Props) {
  if (!loading && entries.length === 0) {
    return (
      <section className="mb-5 shrink-0">
        <h3 className="mb-3 text-sm font-medium text-muted-foreground">Watching</h3>
        <p className="border border-dashed border-border/40 p-8 text-sm text-muted-foreground">
          Nothing on your AniList Watching list yet. Add titles from the Watching tab to track them here.
        </p>
      </section>
    )
  }

  const items = buildWatchingShowItems(entries, localShows)

  return (
    <section className="mb-5 shrink-0">
      <h3 className="mb-3 text-sm font-medium text-muted-foreground">Watching</h3>
      {loading ? (
        <ul
          className="grid grid-cols-[repeat(auto-fill,minmax(min(8.5rem,100%),1fr))] gap-3 p-1"
          aria-busy="true"
          aria-label="Loading Watching list"
        >
          {Array.from({length: 4}, (_, index) => (
            <li key={index}>
              <Skeleton className="aspect-square w-full animate-pulse" />
            </li>
          ))}
        </ul>
      ) : (
        <ul className="grid grid-cols-[repeat(auto-fill,minmax(min(8.5rem,100%),1fr))] gap-3 p-1">
          {items.map((item) => (
            <li key={item.key}>
              <WatchingPoster
                item={item}
                active={item.key === highlightedKey}
                onOpenShow={onOpenShow}
                onFindTorrent={onFindTorrent}
              />
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
