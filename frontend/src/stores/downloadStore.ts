import {create} from 'zustand'
import {DownloadHistory} from '../../wailsjs/go/main/App'
import type {DownloadGroup} from '../lib/downloadGroups'
import type {DownloadView} from '../lib/types'

type DownloadState = {
  jobs: DownloadView[]
  activeTab: DownloadGroup
  setActiveTab: (tab: DownloadGroup) => void
  setJobs: (jobs: DownloadView[]) => void
  upsertJob: (job: DownloadView) => void
  loadHistory: () => Promise<void>
}

export const useDownloadStore = create<DownloadState>((set) => ({
  jobs: [],
  activeTab: 'downloading',

  setActiveTab: (tab) => set({activeTab: tab}),

  setJobs: (jobs) => set({jobs}),

  upsertJob: (job) => {
    set((state) => {
      const exists = state.jobs.some((existing) => existing.id === job.id)
      if (!exists) {
        return {jobs: [job, ...state.jobs]}
      }
      return {
        jobs: state.jobs.map((existing) =>
          existing.id === job.id ? job : existing,
        ),
      }
    })
  },

  loadHistory: async () => {
    const history = await DownloadHistory()
    set({jobs: history ?? []})
  },
}))
