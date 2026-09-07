import {useState} from 'react'
import {AiringEpisodePoster} from './AiringEpisodePoster'
import {AnimeDetailDialog} from './AnimeDetailDialog'
import {IconClose} from './Icons'
import {useSearchStore} from '../stores/searchStore'
import {Button} from '@/components/ui/button'
import {Input} from '@/components/ui/input'

type Props = {
  notice: (msg: string, isError?: boolean) => void
}

const listStatusLabels: Record<string, string> = {
  CURRENT: 'Watching',
  COMPLETED: 'Completed',
  PLANNING: 'Planning',
  PAUSED: 'Paused',
  DROPPED: 'Dropped',
  REPEATING: 'Repeating',
}

export function AnimeSearch({notice}: Props) {
  const searchQuery = useSearchStore((state) => state.animeQuery)
  const searchResults = useSearchStore((state) => state.animeResults)
  const searching = useSearchStore((state) => state.animeLoading)
  const searchError = useSearchStore((state) => state.animeError)
  const setSearchQuery = useSearchStore((state) => state.setAnimeQuery)
  const clearSearch = useSearchStore((state) => state.clearAnimeSearch)
  const searchAnime = useSearchStore((state) => state.searchAnime)
  const updateAnimeListStatus = useSearchStore(
    (state) => state.updateAnimeListStatus,
  )
  const [selectedAnime, setSelectedAnime] = useState<
    (typeof searchResults)[number] | null
  >(null)

  const canClearSearch =
    searchQuery.trim() !== '' || searchResults.length > 0 || searchError !== ''

  return (
    <section
      id="search-anime-panel"
      role="tabpanel"
      aria-labelledby="search-anime-tab"
      className="flex min-h-0 flex-1 flex-col gap-4"
    >
      <div className="shrink-0">
        <p className="text-sm text-muted-foreground">
          Search AniList, then add a title to one of your lists.
        </p>
        <form
          className="mt-4 flex flex-wrap gap-2 bg-card p-4"
          onSubmit={(event) => {
            event.preventDefault()
            void searchAnime()
          }}
        >
          <label className="sr-only" htmlFor="anime-search-query">
            Search AniList
          </label>
          <div className="relative min-w-0 flex-1">
            <Input
              id="anime-search-query"
              value={searchQuery}
              onChange={(event) => setSearchQuery(event.target.value)}
              className="pr-11"
              placeholder="Search AniList by title"
              autoFocus
            />
            {canClearSearch && (
              <Button
                type="button"
                variant="ghost"
                className="absolute top-0 right-0 h-11 min-h-11 w-11 p-0"
                aria-label="Clear anime search"
                onClick={() => clearSearch()}
                disabled={searching}
              >
                <IconClose className="size-4" />
              </Button>
            )}
          </div>
          <Button type="submit" disabled={searching || !searchQuery.trim()}>
            {searching ? 'Searching…' : 'Search'}
          </Button>
        </form>
      </div>

      <div className="min-h-0 flex-1 overflow-auto">
        {searchError && (
          <p className="text-sm text-destructive">{searchError}</p>
        )}

        {searchResults.length > 0 && (
          <section
            className="mt-4 border-t border-border pt-4"
            aria-label="AniList search results"
          >
            <p className="text-sm tabular-nums text-muted-foreground">
              {searchResults.length} results
            </p>
            <ul className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {searchResults.map((anime) => {
                const title = anime.titleEnglish || anime.titleRomaji
                const listCaption = anime.listStatus
                  ? (listStatusLabels[anime.listStatus] ?? anime.listStatus)
                  : 'Not on your list'
                return (
                  <li key={anime.id}>
                    <button
                      type="button"
                      className="flex w-full items-start gap-3 text-left transition-opacity hover:opacity-80 motion-reduce:transition-none"
                      onClick={() => setSelectedAnime(anime)}
                    >
                      <AiringEpisodePoster
                        coverImage={anime.coverImage}
                        size="xxlarge"
                      />
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm font-medium">{title}</p>
                        <p className="text-xs text-muted-foreground">
                          {listCaption}
                        </p>
                        {anime.totalEpisodes > 0 && (
                          <p className="text-xs text-muted-foreground">
                            {anime.totalEpisodes} episodes
                          </p>
                        )}
                      </div>
                    </button>
                  </li>
                )
              })}
            </ul>
          </section>
        )}
      </div>

      {selectedAnime && (
        <AnimeDetailDialog
          anime={selectedAnime}
          notice={notice}
          onClose={() => setSelectedAnime(null)}
          onStatusChange={(status) =>
            updateAnimeListStatus(selectedAnime.id, status)
          }
        />
      )}
    </section>
  )
}
