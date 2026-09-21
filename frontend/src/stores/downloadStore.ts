import {create} from 'zustand'
import {DownloadHistory} from '../../wailsjs/go/main/App'
import {
  downloadGroup,
  indexDownloadJobs,
  type DownloadGroup,
  type DownloadIdsByGroup,
} from '../lib/downloadGroups'
import type {DownloadView} from '../lib/types'

type DownloadState = {
  jobsById: Record<number, DownloadView>
  idsByGroup: DownloadIdsByGroup
  activeTab: DownloadGroup
  setActiveTab: (tab: DownloadGroup) => void
  setJobs: (jobs: DownloadView[]) => void
  upsertJob: (job: DownloadView) => void
  loadHistory: () => Promise<void>
}

const emptyIndex = indexDownloadJobs([])

export const useDownloadStore = create<DownloadState>((set) => ({
  jobsById: emptyIndex.jobsById,
  idsByGroup: emptyIndex.idsByGroup,
  activeTab: 'downloading',

  setActiveTab: (tab) => set({activeTab: tab}),

  setJobs: (jobs) => set(indexDownloadJobs(jobs)),

  upsertJob: (job) => {
    set((state) => {
      const previous = state.jobsById[job.id]
      const jobsById = {...state.jobsById, [job.id]: job}
      if (previous === undefined) {
        const group = downloadGroup(job.status)
        return {
          jobsById,
          idsByGroup: {
            ...state.idsByGroup,
            [group]: [job.id, ...state.idsByGroup[group]],
          },
        }
      }
      const previousGroup = downloadGroup(previous.status)
      const nextGroup = downloadGroup(job.status)
      if (previousGroup === nextGroup) {
        return {jobsById}
      }
      return {
        jobsById,
        idsByGroup: {
          ...state.idsByGroup,
          [previousGroup]: state.idsByGroup[previousGroup].filter(
            (id) => id !== job.id,
          ),
          [nextGroup]: [job.id, ...state.idsByGroup[nextGroup]],
        },
      }
    })
  },

  loadHistory: async () => {
    const history = await DownloadHistory()
    set(indexDownloadJobs(history ?? []))
  },
}))
