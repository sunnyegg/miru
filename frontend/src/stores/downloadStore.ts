import {create} from 'zustand'
import {DownloadHistory} from '../../wailsjs/go/main/App'
import {downloadGroup, type DownloadGroup} from '../lib/downloadGroups'
import type {DownloadView} from '../lib/types'

type DownloadState = {
  jobsById: Record<number, DownloadView>
  groups: Record<DownloadGroup, number[]>
  activeTab: DownloadGroup
  setActiveTab: (tab: DownloadGroup) => void
  upsertJob: (job: DownloadView) => void
  loadHistory: () => Promise<void>
}

export const useDownloadStore = create<DownloadState>((set) => ({
  jobsById: {},
  groups: {downloading: [], seeding: [], completed: []},
  activeTab: 'downloading',

  setActiveTab: (tab) => set({activeTab: tab}),

  upsertJob: (job) => {
    set((state) => {
      const current = state.jobsById[job.id]
      if (current === undefined) {
        const group = downloadGroup(job.status)
        return {
          jobsById: {...state.jobsById, [job.id]: job},
          groups: {...state.groups, [group]: [job.id, ...state.groups[group]]},
        }
      }
      const currentFiles = current.files ?? []
      const nextFiles = job.files ?? []
      const filesEqual =
        currentFiles.length === nextFiles.length &&
        currentFiles.every(
          (file, index) =>
            file.path === nextFiles[index].path &&
            file.length === nextFiles[index].length &&
            file.bytesCompleted === nextFiles[index].bytesCompleted &&
            file.selected === nextFiles[index].selected,
        )
      if (
        current.status === job.status &&
        current.name === job.name &&
        current.percent === job.percent &&
        current.bytesCompleted === job.bytesCompleted &&
        current.bytesTotal === job.bytesTotal &&
        current.bytesUploaded === job.bytesUploaded &&
        current.uploadRatio === job.uploadRatio &&
        current.speedBytesPerSecond === job.speedBytesPerSecond &&
        current.uploadSpeedBytesPerSecond === job.uploadSpeedBytesPerSecond &&
        current.error === job.error &&
        current.source === job.source &&
        current.live === job.live &&
        filesEqual
      ) {
        return state
      }
      const jobsById = {...state.jobsById, [job.id]: job}
      const oldGroup = downloadGroup(current.status)
      const newGroup = downloadGroup(job.status)
      if (oldGroup === newGroup) {
        return {jobsById}
      }
      return {
        jobsById,
        groups: {
          ...state.groups,
          [oldGroup]: state.groups[oldGroup].filter((id) => id !== job.id),
          [newGroup]: [job.id, ...state.groups[newGroup]],
        },
      }
    })
  },

  loadHistory: async () => {
    const history = await DownloadHistory()
    const jobsById: Record<number, DownloadView> = {}
    const groups: Record<DownloadGroup, number[]> = {
      downloading: [],
      seeding: [],
      completed: [],
    }
    for (const job of history ?? []) {
      jobsById[job.id] = job
      groups[downloadGroup(job.status)].push(job.id)
    }
    set({jobsById, groups})
  },
}))
