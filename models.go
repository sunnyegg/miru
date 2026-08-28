package main

type SettingsView struct {
	MpvPath                string  `json:"mpvPath"`
	Anime4KEnabled         bool    `json:"anime4kEnabled"`
	Anime4KShadersReady    bool    `json:"anime4kShadersReady"`
	DownloadDir            string  `json:"downloadDir"`
	SyncThreshold          float64 `json:"syncThreshold"`
	DownloadRateLimit      int64   `json:"downloadRateLimit"`
	UploadRateLimit        int64   `json:"uploadRateLimit"`
	MaxConcurrentDownloads int     `json:"maxConcurrentDownloads"`
	SeedRatio              float64 `json:"seedRatio"`
	NetworkMode            string  `json:"networkMode"`
	Socks5Address          string  `json:"socks5Address"`
	HttpProxyURL           string  `json:"httpProxyUrl"`
	UpdateChannel          string  `json:"updateChannel"`
	DownloadNotifications  bool    `json:"downloadNotifications"`
}

type EpisodeView struct {
	ID                int64  `json:"id"`
	AnilistID         int    `json:"anilistId"`
	EpisodeNumber     int    `json:"episodeNumber"`
	FilePath          string `json:"filePath"`
	DisplayTitle      string `json:"displayTitle"`
	Status            string `json:"status"`
	AnimeTitle        string `json:"animeTitle"`
	CoverImage        string `json:"coverImage"`
	Bound             bool   `json:"bound"`
	Progress          int    `json:"progress"`
	TotalEpisodes     int    `json:"totalEpisodes"`
	MediaStatus       string `json:"mediaStatus"`
	NextAiringEpisode int    `json:"nextAiringEpisode"`
}

type AnimeView struct {
	ID            int    `json:"id"`
	TitleRomaji   string `json:"titleRomaji"`
	TitleEnglish  string `json:"titleEnglish"`
	CoverImage    string `json:"coverImage"`
	TotalEpisodes int    `json:"totalEpisodes"`
	Status        string `json:"status"`
	Synopsis      string `json:"synopsis"`
	ListStatus    string `json:"listStatus"`
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

type FuzzyDateView struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

type AnimeListEntryInput struct {
	MediaID        int    `json:"mediaId"`
	Status         string `json:"status"`
	Progress       int    `json:"progress"`
	ScoreRaw       int    `json:"scoreRaw"`
	Notes          string `json:"notes"`
	Repeat         int    `json:"repeat"`
	Private        bool   `json:"private"`
	StartedYear    int    `json:"startedYear"`
	StartedMonth   int    `json:"startedMonth"`
	StartedDay     int    `json:"startedDay"`
	CompletedYear  int    `json:"completedYear"`
	CompletedMonth int    `json:"completedMonth"`
	CompletedDay   int    `json:"completedDay"`
}

type WatchingEntryView struct {
	MediaID           int           `json:"mediaId"`
	ListStatus        string        `json:"listStatus"`
	Progress          int           `json:"progress"`
	ScoreRaw          int           `json:"scoreRaw"`
	Notes             string        `json:"notes"`
	Repeat            int           `json:"repeat"`
	Private           bool          `json:"private"`
	StartedAt         FuzzyDateView `json:"startedAt"`
	CompletedAt       FuzzyDateView `json:"completedAt"`
	TitleRomaji       string        `json:"titleRomaji"`
	TitleEnglish      string        `json:"titleEnglish"`
	CoverImage        string        `json:"coverImage"`
	TotalEpisodes     int           `json:"totalEpisodes"`
	MediaStatus       string        `json:"mediaStatus"`
	NextAiringEpisode int           `json:"nextAiringEpisode"`
}

type NyaaResultView struct {
	Title     string `json:"title"`
	Link      string `json:"link"`
	Magnet    string `json:"magnet"`
	Published string `json:"published"`
	Size      string `json:"size"`
	Seeders   int    `json:"seeders"`
	Leechers  int    `json:"leechers"`
	Downloads int    `json:"downloads"`
	Trusted   bool   `json:"trusted"`
	Remake    bool   `json:"remake"`
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

type UpdateInfo struct {
	Current    string `json:"current"`
	Latest     string `json:"latest"`
	Available  bool   `json:"available"`
	Notes      string `json:"notes"`
	ReleaseURL string `json:"releaseUrl"`
	AssetName  string `json:"assetName"`
}
