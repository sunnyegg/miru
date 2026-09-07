import {create} from 'zustand'
import {
  ListAnimeList,
  ListAnimeListCounts,
  SaveAnimeListEntry,
  SetAnimeListStatus,
} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import type {AnimeListEntryInput, WatchingEntryView} from '../lib/types'

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
  selectFilter: (filter: ListFilter) => Promise<void>
  loadList: (filter?: ListFilter) => Promise<void>
  loadCounts: () => Promise<void>
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

  setListStatus: async (mediaId, status, totalEpisodes, notice) => {
    try {
      await SetAnimeListStatus(mediaId, status, totalEpisodes)
      notice(quickAddNotices[status])
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
