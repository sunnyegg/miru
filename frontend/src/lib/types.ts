export type TabId = 'library' | 'watching' | 'downloads' | 'calendar' | 'settings'

export type SettingsView = {
  mpvPath: string
  downloadDir: string
  syncThreshold: number
  anilistClientId: string
  downloadRateLimit: number
  uploadRateLimit: number
}

export type EpisodeView = {
  id: number
  anilistId: number
  episodeNumber: number
  filePath: string
  displayTitle: string
  status: string
  animeTitle: string
  coverImage: string
  bound: boolean
}

export type AnimeView = {
  id: number
  titleRomaji: string
  titleEnglish: string
  coverImage: string
  totalEpisodes: number
  status: string
  synopsis: string
}

export type ImportResult = {
  episode: EpisodeView
  candidates: AnimeView[]
  autoBound: boolean
}

export type AnilistStatus = {
  connected: boolean
  username: string
}

export type AiringScheduleView = {
  id: number
  airingAt: number
  episode: number
  mediaId: number
  titleRomaji: string
  titleEnglish: string
  coverImage: string
}

export type WatchingEntryView = {
  mediaId: number
  progress: number
  titleRomaji: string
  titleEnglish: string
  coverImage: string
  totalEpisodes: number
  mediaStatus: string
}

export type DownloadView = {
  id: number
  name: string
  status: string
  bytesCompleted: number
  bytesTotal: number
  bytesUploaded: number
  percent: number
  uploadRatio: number
  speedBytesPerSecond: number
  uploadSpeedBytesPerSecond: number
  error: string
  source: string
}

export type PlaybackEvent = {
  episodeId: number
  percent: number
}

export type SyncEvent = {
  episodeId: number
  ok: boolean
  message: string
}
