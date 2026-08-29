import {create} from 'zustand'
import {persist} from 'zustand/middleware'
import {SearchNyaa, SearchTokyoToshokan} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import type {NyaaResultView} from '../lib/types'

export type SearchSource = 'nyaa' | 'tokyotosho'
export type SearchMode = 'search' | 'feeds'

type NoticeFn = (message: string, isError?: boolean) => void

type SearchState = {
  mode: SearchMode
  query: string
  submittedQuery: string
  source: SearchSource
  results: NyaaResultView[]
  page: number
  loading: boolean
  error: string

  setMode: (mode: SearchMode) => void
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
      mode: 'search',
      query: '',
      submittedQuery: '',
      source: 'nyaa',
      results: [],
      page: 1,
      loading: false,
      error: '',

      setMode: (mode) => set({mode}),
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
        set({mode: 'search', query: trimmed})
        await get().runSearch(notice, trimmed)
      },
    }),
    {
      name: 'miru.search',
      partialize: (state) => ({
        mode: state.mode,
        query: state.query,
        submittedQuery: state.submittedQuery,
        source: state.source,
        results: state.results,
        page: state.page,
      }),
    },
  ),
)
