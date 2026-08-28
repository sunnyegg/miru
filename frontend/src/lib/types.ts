export type TabId = 'library' | 'watching' | 'search' | 'downloads' | 'calendar' | 'settings'

export type SettingsView = {
  mpvPath: string
  anime4kEnabled: boolean
  anime4kShadersReady: boolean
  downloadDir: string
  syncThreshold: number
  downloadRateLimit: number
  uploadRateLimit: number
  maxConcurrentDownloads: number
  seedRatio: number
  networkMode: string
  socks5Address: string
  httpProxyUrl: string
  updateChannel: string
  rssPollIntervalMinutes: number
  discordRpcEnabled: boolean
  discordAppId: string
  downloadNotifications: boolean
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
  progress: number
  totalEpisodes: number
  mediaStatus: string
  nextAiringEpisode: number
}

export type AnimeView = {
  id: number
  titleRomaji: string
  titleEnglish: string
  coverImage: string
  totalEpisodes: number
  status: string
  synopsis: string
  listStatus: string
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

export type FuzzyDateView = {
  year: number
  month: number
  day: number
}

export type AnimeListEntryInput = {
  mediaId: number
  status: string
  progress: number
  scoreRaw: number
  notes: string
  repeat: number
  private: boolean
  startedYear: number
  startedMonth: number
  startedDay: number
  completedYear: number
  completedMonth: number
  completedDay: number
}

export type WatchingEntryView = {
  mediaId: number
  listStatus: string
  progress: number
  scoreRaw: number
  notes: string
  repeat: number
  private: boolean
  startedAt: FuzzyDateView
  completedAt: FuzzyDateView
  titleRomaji: string
  titleEnglish: string
  coverImage: string
  totalEpisodes: number
  mediaStatus: string
  nextAiringEpisode: number
}

export type NyaaResultView = {
  title: string
  link: string
  magnet: string
  published: string
  size: string
  seeders: number
  leechers: number
  downloads: number
  trusted: boolean
  remake: boolean
}

export type RSSFeedView = {
  id: number
  url: string
  title: string
  enabled: boolean
  lastPolled: string
  newCount: number
}

export type RSSFeedItemView = {
  id: number
  feedId: number
  feedTitle: string
  title: string
  link: string
  magnet: string
  published: string
  isNew: boolean
}

export type TorrentFileView = {
  path: string
  length: number
  bytesCompleted: number
  selected: boolean
  isVideo: boolean
}

export type TorrentContentsView = {
  name: string
  bytesTotal: number
  files: TorrentFileView[]
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
  live: boolean
  files: TorrentFileView[]
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

export type UpdateInfo = {
  current: string
  latest: string
  available: boolean
  notes: string
  releaseUrl: string
  assetName: string
}
