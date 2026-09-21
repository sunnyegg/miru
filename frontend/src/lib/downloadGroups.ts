export type DownloadGroup = 'downloading' | 'seeding' | 'completed'

export function downloadGroup(status: string): DownloadGroup {
  if (status === 'SEEDING') {
    return 'seeding'
  }
  if (status === 'COMPLETED' || status === 'FAILED' || status === 'CANCELLED') {
    return 'completed'
  }
  return 'downloading'
}
