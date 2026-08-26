package main

type SettingsView struct {
	MpvPath           string  `json:"mpvPath"`
	DownloadDir       string  `json:"downloadDir"`
	SyncThreshold     float64 `json:"syncThreshold"`
	AnilistClientId   string  `json:"anilistClientId"`
	DownloadRateLimit int64   `json:"downloadRateLimit"`
	UploadRateLimit   int64   `json:"uploadRateLimit"`
}

type EpisodeView struct {
	ID            int64  `json:"id"`
	AnilistID     int    `json:"anilistId"`
	EpisodeNumber int    `json:"episodeNumber"`
	FilePath      string `json:"filePath"`
	DisplayTitle  string `json:"displayTitle"`
	Status        string `json:"status"`
	AnimeTitle    string `json:"animeTitle"`
	CoverImage    string `json:"coverImage"`
	Bound         bool   `json:"bound"`
}

type AnimeView struct {
	ID            int    `json:"id"`
	TitleRomaji   string `json:"titleRomaji"`
	TitleEnglish  string `json:"titleEnglish"`
	CoverImage    string `json:"coverImage"`
	TotalEpisodes int    `json:"totalEpisodes"`
	Status        string `json:"status"`
	Synopsis      string `json:"synopsis"`
}

type ImportResult struct {
	Episode    EpisodeView `json:"episode"`
	Candidates []AnimeView `json:"candidates"`
	AutoBound  bool        `json:"autoBound"`
}

type AnilistStatus struct {
	Connected bool   `json:"connected"`
	Username  string `json:"username"`
}

type AiringScheduleView struct {
	ID           int64  `json:"id"`
	AiringAt     int64  `json:"airingAt"`
	Episode      int    `json:"episode"`
	MediaID      int    `json:"mediaId"`
	TitleRomaji  string `json:"titleRomaji"`
	TitleEnglish string `json:"titleEnglish"`
	CoverImage   string `json:"coverImage"`
}

type PlaybackEvent struct {
	EpisodeID int64   `json:"episodeId"`
	Percent   float64 `json:"percent"`
}

type SyncEvent struct {
	EpisodeID int64  `json:"episodeId"`
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
}
