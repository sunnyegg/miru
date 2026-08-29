import {create} from 'zustand'
import {persist} from 'zustand/middleware'
import type {PlaybackEvent} from '../lib/types'

type PlaybackState = {
  playing: PlaybackEvent | null
  lastPlayback: PlaybackEvent | null
  trackProgress: (event: PlaybackEvent) => void
  clearPlaying: () => void
}

export const usePlaybackStore = create<PlaybackState>()(
  persist(
    (set) => ({
      playing: null,
      lastPlayback: null,

      trackProgress: (event) => set({playing: event, lastPlayback: event}),

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
