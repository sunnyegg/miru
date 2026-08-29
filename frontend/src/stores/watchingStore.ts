import {create} from 'zustand'
import {
  ListAnimeList,
  ListAnimeListCounts,
  SaveAnimeListEntry,
  SearchAnime,
  SetAnimeListStatus,
} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import type {
  AnimeListEntryInput,
  AnimeView,
  WatchingEntryView,
} from '../lib/types'

export type ListFilter =
  | 'CURRENT'
  | 'COMPLETED'
  | 'PLANNING'
  | 'PAUSED'
  | 'DROPPED'
  | 'REPEATING'

export type QuickAddStatus = 'CURRENT' | 'PLANNING' | 'COMPLETED'

const quickAddNotices: Record<QuickAddStatus, string> = {
  CURRENT: 'Added to Watching',
  PLANNING: 'Added to Planning',
  COMPLETED: 'Added to Completed',
}

type NoticeFn = (message: string, isError?: boolean) => void

type WatchingState = {
  listFilter: ListFilter
  entries: WatchingEntryView[]
  counts: Partial<Record<ListFilter, number>>
  countsLoading: boolean
  countsError: string
  loading: boolean
  notConnected: boolean
  error: string
  searchQuery: string
  searchResults: AnimeView[]
  searching: boolean
  searchError: string
  setSearchQuery: (query: string) => void
  clearSearch: () => void
  selectFilter: (filter: ListFilter) => Promise<void>
  loadList: (filter?: ListFilter) => Promise<void>
  loadCounts: () => Promise<void>
  searchAnime: () => Promise<void>
  setListStatus: (
    mediaId: number,
    status: QuickAddStatus,
    totalEpisodes: number,
    notice: NoticeFn,
  ) => Promise<boolean>
  saveEntry: (input: AnimeListEntryInput, notice: NoticeFn) => Promise<void>
}

export const useWatchingStore = create<WatchingState>((set, get) => ({
  listFilter: 'CURRENT',
  entries: [],
  counts: {},
  countsLoading: true,
  countsError: '',
  loading: true,
  notConnected: false,
  error: '',
  searchQuery: '',
  searchResults: [],
  searching: false,
  searchError: '',

  setSearchQuery: (query) => set({searchQuery: query}),

  clearSearch: () => set({searchQuery: '', searchResults: [], searchError: ''}),

  loadList: async (filter) => {
    const listFilter = filter ?? get().listFilter
    set({loading: true, notConnected: false, error: ''})
    try {
      const result = await ListAnimeList(listFilter)
      set({entries: result ?? []})
    } catch (err) {
      const message = errorMessage(err)
      if (message === 'AniList not connected') {
        set({notConnected: true, entries: []})
      } else {
        set({error: message})
      }
    } finally {
      set({loading: false})
    }
  },

  loadCounts: async () => {
    set({countsLoading: true, countsError: ''})
    try {
      const counts = await ListAnimeListCounts()
      set({counts: counts ?? {}})
    } catch (err) {
      set({counts: {}, countsError: errorMessage(err)})
    } finally {
      set({countsLoading: false})
    }
  },

  selectFilter: async (filter) => {
    set({listFilter: filter})
    await get().loadList(filter)
  },

  searchAnime: async () => {
    const trimmed = get().searchQuery.trim()
    if (!trimmed) {
      set({searchError: 'Enter an anime title to search.'})
      return
    }
    set({searching: true, searchError: ''})
    try {
      const found = await SearchAnime(trimmed)
      set({searchResults: found ?? []})
    } catch (err) {
      set({searchError: errorMessage(err), searchResults: []})
    } finally {
      set({searching: false})
    }
  },

  setListStatus: async (mediaId, status, totalEpisodes, notice) => {
    try {
      await SetAnimeListStatus(mediaId, status, totalEpisodes)
      notice(quickAddNotices[status])
      set((state) => ({
        searchResults: state.searchResults.map((anime) =>
          anime.id === mediaId ? {...anime, listStatus: status} : anime,
        ),
      }))
      await get().loadList()
      await get().loadCounts()
      return true
    } catch (err) {
      notice(errorMessage(err), true)
      return false
    }
  },

  saveEntry: async (input, notice) => {
    try {
      await SaveAnimeListEntry(input)
      notice('List entry updated')
      await get().loadList()
      await get().loadCounts()
    } catch (err) {
      notice(errorMessage(err), true)
      throw err
    }
  },
}))
