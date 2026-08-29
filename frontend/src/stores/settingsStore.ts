import {create} from 'zustand'

export type SettingsTab =
  | 'desktop'
  | 'playback'
  | 'downloads'
  | 'network'
  | 'anilist'
  | 'about'

type SettingsUIState = {
  activeTab: SettingsTab
  reloadKey: number
  setActiveTab: (tab: SettingsTab) => void
  bumpReloadKey: () => void
}

export const useSettingsStore = create<SettingsUIState>((set) => ({
  activeTab: 'desktop',
  reloadKey: 0,

  setActiveTab: (tab) => set({activeTab: tab}),

  bumpReloadKey: () => set((state) => ({reloadKey: state.reloadKey + 1})),
}))
