import {create} from 'zustand'
import {ListAnimeList, SaveAnimeListEntry, SearchAnime, SetAnimeListStatus} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import type {AnimeListEntryInput, AnimeView, WatchingEntryView} from '../lib/types'

export type ListFilter =
  | 'CURRENT'
  | 'COMPLETED'
  | 'PLANNING'
  | 'PAUSED'
  | 'DROPPED'
  | 'REPEATING'

type NoticeFn = (message: string, isError?: boolean) => void

type WatchingState = {
  listFilter: ListFilter
  entries: WatchingEntryView[]
  loading: boolean
  notConnected: boolean
  error: string
  searchQuery: string
  searchResults: AnimeView[]
  searching: boolean
  searchError: string
  setSearchQuery: (query: string) => void
  selectFilter: (filter: ListFilter) => Promise<void>
  loadList: (filter?: ListFilter) => Promise<void>
  searchAnime: () => Promise<void>
  markWatching: (mediaId: number, notice: NoticeFn) => Promise<void>
  saveEntry: (input: AnimeListEntryInput, notice: NoticeFn) => Promise<void>
}

export const useWatchingStore = create<WatchingState>((set, get) => ({
  listFilter: 'CURRENT',
  entries: [],
  loading: true,
  notConnected: false,
  error: '',
  searchQuery: '',
  searchResults: [],
  searching: false,
  searchError: '',

  setSearchQuery: (query) => set({searchQuery: query}),

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

  markWatching: async (mediaId, notice) => {
    try {
      await SetAnimeListStatus(mediaId, 'CURRENT', 0)
      notice('Added to Watching')
      set((state) => ({
        searchResults: state.searchResults.map((anime) =>
          anime.id === mediaId ? {...anime, listStatus: 'CURRENT'} : anime
        ),
      }))
      await get().loadList()
    } catch (err) {
      notice(errorMessage(err), true)
    }
  },

  saveEntry: async (input, notice) => {
    try {
      await SaveAnimeListEntry(input)
      notice('List entry updated')
      await get().loadList()
    } catch (err) {
      notice(errorMessage(err), true)
      throw err
    }
  },
}))
