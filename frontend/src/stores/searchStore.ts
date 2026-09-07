import {create} from 'zustand'
import {persist} from 'zustand/middleware'
import {
  SearchAnime,
  SearchNyaa,
  SearchTokyoToshokan,
} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import type {AnimeView, NyaaResultView} from '../lib/types'

export type SearchSource = 'nyaa' | 'tokyotosho'
export type SearchTab = 'anime' | 'torrent'
export type TorrentMode = 'search' | 'feeds'

type NoticeFn = (message: string, isError?: boolean) => void

type SearchState = {
  tab: SearchTab
  torrentMode: TorrentMode
  animeQuery: string
  animeResults: AnimeView[]
  animeLoading: boolean
  animeError: string
  query: string
  submittedQuery: string
  source: SearchSource
  results: NyaaResultView[]
  page: number
  loading: boolean
  error: string

  setTab: (tab: SearchTab) => void
  setTorrentMode: (mode: TorrentMode) => void
  setAnimeQuery: (query: string) => void
  clearAnimeSearch: () => void
  searchAnime: () => Promise<void>
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
      animeQuery: '',
      animeResults: [],
      animeLoading: false,
      animeError: '',
      query: '',
      submittedQuery: '',
      source: 'nyaa',
      results: [],
      page: 1,
      loading: false,
      error: '',

      setTab: (tab) => set({tab}),
      setTorrentMode: (torrentMode) => set({torrentMode}),
      setAnimeQuery: (animeQuery) => set({animeQuery}),
      clearAnimeSearch: () =>
        set({animeQuery: '', animeResults: [], animeError: ''}),

      searchAnime: async () => {
        const trimmed = get().animeQuery.trim()
        if (!trimmed) {
          set({animeError: 'Enter an anime title to search.'})
          return
        }
        set({animeLoading: true, animeError: ''})
        try {
          const found = await SearchAnime(trimmed)
          set({animeResults: found ?? []})
        } catch (err) {
          set({animeError: errorMessage(err), animeResults: []})
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

        set({loading: true, error: '', submittedQuery: trimmed})
        try {
          const found =
            source === 'tokyotosho'
              ? await SearchTokyoToshokan(trimmed)
              : await SearchNyaa(trimmed)
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
