import type {DownloadView} from './types'

export type DownloadGroup = 'downloading' | 'seeding' | 'completed'

export type GroupedDownloads = Record<DownloadGroup, DownloadView[]>

export function downloadGroup(status: string): DownloadGroup {
  if (status === 'SEEDING') {
    return 'seeding'
  }
  if (status === 'COMPLETED' || status === 'FAILED' || status === 'CANCELLED') {
    return 'completed'
  }
  return 'downloading'
}

export function groupDownloads(jobs: DownloadView[]): GroupedDownloads {
  const grouped: GroupedDownloads = {
    downloading: [],
    seeding: [],
    completed: [],
  }

  for (const job of jobs) {
    grouped[downloadGroup(job.status)].push(job)
  }

  return grouped
}
