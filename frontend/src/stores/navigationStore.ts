import {create} from 'zustand'
import type {TabId} from '../lib/types'

type NavigationState = {
  tab: TabId
  setTab: (tab: TabId) => void
}

export const useNavigationStore = create<NavigationState>((set) => ({
  tab: 'library',
  setTab: (tab) => set({tab}),
}))
