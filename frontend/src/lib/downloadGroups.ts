import type {DownloadView} from './types'

export type DownloadGroup = 'downloading' | 'seeding' | 'completed'

export type DownloadIdsByGroup = Record<DownloadGroup, number[]>

export function downloadGroup(status: string): DownloadGroup {
  if (status === 'SEEDING') {
    return 'seeding'
  }
  if (status === 'COMPLETED' || status === 'FAILED' || status === 'CANCELLED') {
    return 'completed'
  }
  return 'downloading'
}

export function indexDownloadJobs(jobs: DownloadView[]): {
  jobsById: Record<number, DownloadView>
  idsByGroup: DownloadIdsByGroup
} {
  const jobsById: Record<number, DownloadView> = {}
  const idsByGroup: DownloadIdsByGroup = {
    downloading: [],
    seeding: [],
    completed: [],
  }
  for (const job of jobs) {
    jobsById[job.id] = job
    idsByGroup[downloadGroup(job.status)].push(job.id)
  }
  return {jobsById, idsByGroup}
}
