import {create} from 'zustand'
import {ListAnimeList, ListEpisodes} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import type {EpisodeView, WatchingEntryView} from '../lib/types'

type NoticeFn = (message: string, isError?: boolean) => void

type LibraryState = {
  episodes: EpisodeView[]
  watchingEntries: WatchingEntryView[]
  selectedKey: string | null
  loading: boolean
  watchingLoading: boolean
  loadError: string
  setSelectedKey: (key: string | null) => void
  reload: (notice?: NoticeFn) => Promise<void>
  reloadWatching: () => Promise<void>
}

export const useLibraryStore = create<LibraryState>((set, get) => ({
  episodes: [],
  watchingEntries: [],
  selectedKey: null,
  loading: true,
  watchingLoading: true,
  loadError: '',

  setSelectedKey: (key) => set({selectedKey: key}),

  reload: async (notice) => {
    set({loadError: ''})
    try {
      const rows = await ListEpisodes()
      set({episodes: rows ?? []})
    } catch (err) {
      const message = errorMessage(err)
      set({loadError: message})
      if (get().episodes.length > 0 && notice) {
        notice(message, true)
      }
    } finally {
      set({loading: false})
    }
  },

  reloadWatching: async () => {
    set({watchingLoading: true})
    try {
      const result = await ListAnimeList('CURRENT')
      set({watchingEntries: result ?? []})
    } catch {
      set({watchingEntries: []})
    } finally {
      set({watchingLoading: false})
    }
  },
}))
