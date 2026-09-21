import {create} from 'zustand'
import {persist} from 'zustand/middleware'
import type {PlaybackEvent} from '../lib/types'

type PlaybackState = {
  playing: PlaybackEvent | null
  lastPlayback: PlaybackEvent | null
  progressByEpisodeId: Record<number, number>
  trackProgress: (event: PlaybackEvent) => void
  clearPlaying: () => void
}

export const usePlaybackStore = create<PlaybackState>()(
  persist(
    (set) => ({
      playing: null,
      lastPlayback: null,
      progressByEpisodeId: {},

      trackProgress: (event) =>
        set((state) => {
          if (
            state.playing?.episodeId === event.episodeId &&
            state.playing.percent === event.percent
          ) {
            return state
          }
          return {
            playing: event,
            lastPlayback: event,
            progressByEpisodeId:
              state.progressByEpisodeId[event.episodeId] === event.percent
                ? state.progressByEpisodeId
                : {
                    ...state.progressByEpisodeId,
                    [event.episodeId]: event.percent,
                  },
          }
        }),

      clearPlaying: () => set({playing: null}),
    }),
    {
      name: 'miru.playback',
      partialize: (state) => ({
        lastPlayback: state.lastPlayback,
      }),
    },
  ),
)
