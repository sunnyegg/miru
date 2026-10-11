import {memo, useState} from 'react'
import {AiringEpisodePoster} from './AiringEpisodePoster'
import {AnimeDetailDialog} from './AnimeDetailDialog'
import {IconClose} from './Icons'
import {mediaSeasons, seasonDisplayName, type MediaSeason} from '../lib/season'
import type {AnimeView} from '../lib/types'
import {useSearchStore} from '../stores/searchStore'
import {Button} from '@/components/ui/button'
import {Input} from '@/components/ui/input'
import {NativeSelect, NativeSelectOption} from '@/components/ui/native-select'

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

type AnimeSearchResultProps = {
  anime: AnimeView
  onSelect: (anime: AnimeView) => void
}

const AnimeSearchResult = memo(function AnimeSearchResult({
  anime,
  onSelect,
}: AnimeSearchResultProps) {
  const title = anime.titleEnglish || anime.titleRomaji
  const listCaption = anime.listStatus
    ? (listStatusLabels[anime.listStatus] ?? anime.listStatus)
    : 'Not on your list'

  return (
    <li>
      <button
        type="button"
        className="flex w-full items-start gap-3 text-left transition-opacity hover:opacity-80 motion-reduce:transition-none"
        onClick={() => onSelect(anime)}
      >
        <AiringEpisodePoster coverImage={anime.coverImage} size="xxlarge" />
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium">{title}</p>
          <p className="text-xs text-muted-foreground">{listCaption}</p>
          {anime.totalEpisodes > 0 && (
            <p className="text-xs text-muted-foreground">
              {anime.totalEpisodes} episodes
            </p>
          )}
        </div>
      </button>
    </li>
  )
})

export function AnimeSearch({notice}: Props) {
  const animeMode = useSearchStore((state) => state.animeMode)
  const searchQuery = useSearchStore((state) => state.animeQuery)
  const searchResults = useSearchStore((state) => state.animeResults)
  const searching = useSearchStore((state) => state.animeLoading)
  const searchError = useSearchStore((state) => state.animeError)
  const animeSeason = useSearchStore((state) => state.animeSeason)
  const animeSeasonYear = useSearchStore((state) => state.animeSeasonYear)
  const animeSeasonHasNext = useSearchStore((state) => state.animeSeasonHasNext)
  const setAnimeMode = useSearchStore((state) => state.setAnimeMode)
  const setSearchQuery = useSearchStore((state) => state.setAnimeQuery)
  const setAnimeSeason = useSearchStore((state) => state.setAnimeSeason)
  const setAnimeSeasonYear = useSearchStore((state) => state.setAnimeSeasonYear)
  const clearSearch = useSearchStore((state) => state.clearAnimeSearch)
  const searchAnime = useSearchStore((state) => state.searchAnime)
  const browseAnimeSeason = useSearchStore((state) => state.browseAnimeSeason)
  const loadMoreAnimeSeason = useSearchStore(
    (state) => state.loadMoreAnimeSeason,
  )
  const updateAnimeListStatus = useSearchStore(
    (state) => state.updateAnimeListStatus,
  )
  const [selectedAnime, setSelectedAnime] = useState<
    (typeof searchResults)[number] | null
  >(null)

  const canClearSearch =
    searchQuery.trim() !== '' || searchResults.length > 0 || searchError !== ''

  const yearOptions: number[] = []
  const currentYear = new Date().getFullYear()
  for (let year = currentYear + 1; year >= 1990; year -= 1) {
    yearOptions.push(year)
  }

  return (
    <section
      id="search-anime-panel"
      role="tabpanel"
      aria-labelledby="search-anime-tab"
      className="flex min-h-0 flex-1 flex-col gap-4"
    >
      <div className="shrink-0">
        <p className="text-sm text-muted-foreground">
          Search AniList by title, or browse a season, then add a title to one
          of your lists.
        </p>
        <div
          className="mt-4 flex flex-wrap gap-2"
          role="group"
          aria-label="Anime search mode"
        >
          <Button
            type="button"
            variant={animeMode === 'title' ? 'default' : 'muted'}
            aria-pressed={animeMode === 'title'}
            onClick={() => setAnimeMode('title')}
          >
            Title
          </Button>
          <Button
            type="button"
            variant={animeMode === 'season' ? 'default' : 'muted'}
            aria-pressed={animeMode === 'season'}
            onClick={() => setAnimeMode('season')}
          >
            Season
          </Button>
        </div>

        {animeMode === 'title' ? (
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
        ) : (
          <form
            className="mt-4 flex flex-wrap gap-2 bg-card p-4"
            onSubmit={(event) => {
              event.preventDefault()
              void browseAnimeSeason()
            }}
          >
            <label className="sr-only" htmlFor="anime-season">
              Season
            </label>
            <NativeSelect
              id="anime-season"
              className="w-auto min-w-36 shrink-0"
              value={animeSeason}
              onChange={(event) =>
                setAnimeSeason(event.target.value as MediaSeason)
              }
            >
              {mediaSeasons.map((season) => (
                <NativeSelectOption key={season} value={season}>
                  {seasonDisplayName(season)}
                </NativeSelectOption>
              ))}
            </NativeSelect>
            <label className="sr-only" htmlFor="anime-season-year">
              Year
            </label>
            <NativeSelect
              id="anime-season-year"
              className="w-auto min-w-28 shrink-0"
              value={String(animeSeasonYear)}
              onChange={(event) =>
                setAnimeSeasonYear(Number(event.target.value))
              }
            >
              {yearOptions.map((year) => (
                <NativeSelectOption key={year} value={String(year)}>
                  {year}
                </NativeSelectOption>
              ))}
            </NativeSelect>
            <Button type="submit" disabled={searching}>
              {searching && searchResults.length === 0 ? 'Loading…' : 'Browse'}
            </Button>
          </form>
        )}
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
              {searchResults.map((anime) => (
                <AnimeSearchResult
                  key={anime.id}
                  anime={anime}
                  onSelect={setSelectedAnime}
                />
              ))}
            </ul>
            {animeMode === 'season' && animeSeasonHasNext && (
              <div className="mt-4">
                <Button
                  type="button"
                  variant="secondary"
                  disabled={searching}
                  onClick={() => void loadMoreAnimeSeason()}
                >
                  {searching ? 'Loading…' : 'Load more'}
                </Button>
              </div>
            )}
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
