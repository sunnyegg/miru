import {create} from 'zustand'
import {
  ListAnimeList,
  ListEpisodes,
  ListStreamingEpisodeThumbnails,
} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import type {
  EpisodeView,
  StreamingEpisodeThumbnailView,
  WatchingEntryView,
} from '../lib/types'

type NoticeFn = (message: string, isError?: boolean) => void

type LibraryState = {
  episodes: EpisodeView[]
  watchingEntries: WatchingEntryView[]
  episodeThumbnailsByMediaId: Record<number, Record<number, string>>
  selectedKey: string | null
  loading: boolean
  watchingLoading: boolean
  loadError: string
  episodeThumbnailsError: string
  setSelectedKey: (key: string | null) => void
  reload: (notice?: NoticeFn) => Promise<void>
  reloadWatching: () => Promise<void>
  loadEpisodeThumbnails: (mediaId: number) => Promise<void>
}

const thumbnailLoads = new Map<number, Promise<void>>()

function mapEpisodeThumbnails(
  rows: StreamingEpisodeThumbnailView[],
): Record<number, string> {
  const mapped: Record<number, string> = {}
  for (const row of rows ?? []) {
    if (row.thumbnail) {
      mapped[row.episodeNumber] = row.thumbnail
    }
  }
  return mapped
}

export const useLibraryStore = create<LibraryState>((set, get) => ({
  episodes: [],
  watchingEntries: [],
  episodeThumbnailsByMediaId: {},
  selectedKey: null,
  loading: true,
  watchingLoading: true,
  loadError: '',
  episodeThumbnailsError: '',

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

  loadEpisodeThumbnails: async (mediaId) => {
    if (mediaId <= 0) {
      return
    }
    set({episodeThumbnailsError: ''})
    if (get().episodeThumbnailsByMediaId[mediaId]) {
      return
    }

    const pending = thumbnailLoads.get(mediaId)
    if (pending) {
      await pending
      return
    }

    const request = (async () => {
      try {
        const rows = await ListStreamingEpisodeThumbnails(mediaId)
        set((state) => ({
          episodeThumbnailsByMediaId: {
            ...state.episodeThumbnailsByMediaId,
            [mediaId]: mapEpisodeThumbnails(rows ?? []),
          },
        }))
      } catch (err) {
        set({episodeThumbnailsError: errorMessage(err)})
      } finally {
        thumbnailLoads.delete(mediaId)
      }
    })()
    thumbnailLoads.set(mediaId, request)
    await request
  },
}))
