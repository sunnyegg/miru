import {create} from 'zustand'
import {persist} from 'zustand/middleware'
import {
  BrowseAnimeSeason,
  GetSettings,
  SearchAnime,
  SearchNyaa,
  SearchTokyoToshokan,
} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import {withTorrentSearchPrefix} from '../lib/libraryWatching'
import {currentAniListSeason, type MediaSeason} from '../lib/season'
import type {AnimeView, NyaaResultView} from '../lib/types'

export type SearchSource = 'nyaa' | 'tokyotosho'
export type SearchTab = 'anime' | 'torrent'
export type TorrentMode = 'search' | 'feeds'
export type AnimeSearchMode = 'title' | 'season'

type NoticeFn = (message: string, isError?: boolean) => void

const defaultSeason = currentAniListSeason()

type SearchState = {
  tab: SearchTab
  torrentMode: TorrentMode
  animeMode: AnimeSearchMode
  animeQuery: string
  animeResults: AnimeView[]
  animeLoading: boolean
  animeError: string
  animeSeason: MediaSeason
  animeSeasonYear: number
  animeSeasonPage: number
  animeSeasonHasNext: boolean
  query: string
  submittedQuery: string
  source: SearchSource
  results: NyaaResultView[]
  page: number
  loading: boolean
  error: string

  setTab: (tab: SearchTab) => void
  setTorrentMode: (mode: TorrentMode) => void
  setAnimeMode: (mode: AnimeSearchMode) => void
  setAnimeQuery: (query: string) => void
  setAnimeSeason: (season: MediaSeason) => void
  setAnimeSeasonYear: (year: number) => void
  clearAnimeSearch: () => void
  searchAnime: () => Promise<void>
  browseAnimeSeason: () => Promise<void>
  loadMoreAnimeSeason: () => Promise<void>
  updateAnimeListStatus: (mediaId: number, status: string) => void
  setQuery: (query: string) => void
  setPage: (page: number) => void
  changeSource: (nextSource: SearchSource, notice: NoticeFn) => Promise<void>
  runSearch: (
    notice: NoticeFn,
    searchQuery?: string,
    searchSource?: SearchSource,
  ) => Promise<void>
  prefillSearch: (prefillQuery: string, notice: NoticeFn) => Promise<void>
}

export const useSearchStore = create<SearchState>()(
  persist(
    (set, get) => ({
      tab: 'anime',
      torrentMode: 'search',
      animeMode: 'title',
      animeQuery: '',
      animeResults: [],
      animeLoading: false,
      animeError: '',
      animeSeason: defaultSeason.season,
      animeSeasonYear: defaultSeason.year,
      animeSeasonPage: 0,
      animeSeasonHasNext: false,
      query: '',
      submittedQuery: '',
      source: 'nyaa',
      results: [],
      page: 1,
      loading: false,
      error: '',

      setTab: (tab) => set({tab}),
      setTorrentMode: (torrentMode) => set({torrentMode}),
      setAnimeMode: (animeMode) => set({animeMode}),
      setAnimeQuery: (animeQuery) => set({animeQuery}),
      setAnimeSeason: (animeSeason) => set({animeSeason}),
      setAnimeSeasonYear: (animeSeasonYear) => set({animeSeasonYear}),
      clearAnimeSearch: () =>
        set({
          animeQuery: '',
          animeResults: [],
          animeError: '',
          animeSeasonPage: 0,
          animeSeasonHasNext: false,
        }),

      searchAnime: async () => {
        const trimmed = get().animeQuery.trim()
        if (!trimmed) {
          set({animeError: 'Enter an anime title to search.'})
          return
        }
        set({animeLoading: true, animeError: '', animeSeasonHasNext: false})
        try {
          const found = await SearchAnime(trimmed)
          set({animeResults: found ?? [], animeSeasonPage: 0})
        } catch (err) {
          set({animeError: errorMessage(err), animeResults: []})
        } finally {
          set({animeLoading: false})
        }
      },

      browseAnimeSeason: async () => {
        const {animeSeason, animeSeasonYear} = get()
        if (animeSeasonYear <= 0) {
          set({animeError: 'Enter a valid season year.'})
          return
        }
        set({animeLoading: true, animeError: ''})
        try {
          const page = await BrowseAnimeSeason(animeSeason, animeSeasonYear, 1)
          set({
            animeResults: page?.items ?? [],
            animeSeasonPage: 1,
            animeSeasonHasNext: Boolean(page?.hasNextPage),
          })
        } catch (err) {
          set({
            animeError: errorMessage(err),
            animeResults: [],
            animeSeasonPage: 0,
            animeSeasonHasNext: false,
          })
        } finally {
          set({animeLoading: false})
        }
      },

      loadMoreAnimeSeason: async () => {
        const {
          animeLoading,
          animeSeasonHasNext,
          animeSeason,
          animeSeasonYear,
          animeSeasonPage,
          animeResults,
        } = get()
        if (animeLoading || !animeSeasonHasNext || animeSeasonPage < 1) {
          return
        }
        const nextPage = animeSeasonPage + 1
        set({animeLoading: true, animeError: ''})
        try {
          const page = await BrowseAnimeSeason(
            animeSeason,
            animeSeasonYear,
            nextPage,
          )
          set({
            animeResults: [...animeResults, ...(page?.items ?? [])],
            animeSeasonPage: nextPage,
            animeSeasonHasNext: Boolean(page?.hasNextPage),
          })
        } catch (err) {
          set({animeError: errorMessage(err)})
        } finally {
          set({animeLoading: false})
        }
      },

      updateAnimeListStatus: (mediaId, status) =>
        set((state) => ({
          animeResults: state.animeResults.map((anime) =>
            anime.id === mediaId ? {...anime, listStatus: status} : anime,
          ),
        })),

      setQuery: (query) => set({query}),
      setPage: (page) => set({page}),

      changeSource: async (nextSource, notice) => {
        set({source: nextSource})
        const {submittedQuery} = get()
        if (submittedQuery) {
          await get().runSearch(notice, submittedQuery, nextSource)
        }
      },

      runSearch: async (notice, searchQuery, searchSource) => {
        const state = get()
        const trimmed = (searchQuery ?? state.query).trim()
        const source = searchSource ?? state.source

        if (!trimmed) {
          set({error: 'Enter an anime title to search.'})
          return
        }

        let searchPrefix = ''
        try {
          const settings = await GetSettings()
          searchPrefix = settings?.torrentSearchPrefix || ''
        } catch {
          // Search without prefix when settings cannot be loaded.
        }
        const queryForIndexer = withTorrentSearchPrefix(trimmed, searchPrefix)

        set({
          loading: true,
          error: '',
          query: queryForIndexer,
          submittedQuery: queryForIndexer,
        })
        try {
          const found =
            source === 'tokyotosho'
              ? await SearchTokyoToshokan(queryForIndexer)
              : await SearchNyaa(queryForIndexer)
          set({results: found ?? [], page: 1})
        } catch (err) {
          const message = errorMessage(err)
          set({error: message})
          notice(message, true)
        } finally {
          set({loading: false})
        }
      },

      prefillSearch: async (prefillQuery, notice) => {
        const trimmed = prefillQuery.trim()
        if (!trimmed) {
          return
        }
        set({tab: 'torrent', torrentMode: 'search', query: trimmed})
        await get().runSearch(notice, trimmed)
      },
    }),
    {
      name: 'miru.search',
      partialize: (state) => ({
        torrentMode: state.torrentMode,
        query: state.query,
        submittedQuery: state.submittedQuery,
        source: state.source,
        results: state.results,
        page: state.page,
      }),
    },
  ),
)
