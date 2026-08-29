export namespace main {
	
	export class AiringScheduleView {
	    id: number;
	    airingAt: number;
	    episode: number;
	    mediaId: number;
	    titleRomaji: string;
	    titleEnglish: string;
	    coverImage: string;
	
	    static createFrom(source: any = {}) {
	        return new AiringScheduleView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.airingAt = source["airingAt"];
	        this.episode = source["episode"];
	        this.mediaId = source["mediaId"];
	        this.titleRomaji = source["titleRomaji"];
	        this.titleEnglish = source["titleEnglish"];
	        this.coverImage = source["coverImage"];
	    }
	}
	export class AnilistStatus {
	    connected: boolean;
	    username: string;
	
	    static createFrom(source: any = {}) {
	        return new AnilistStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connected = source["connected"];
	        this.username = source["username"];
	    }
	}
	export class AnimeListEntryInput {
	    mediaId: number;
	    status: string;
	    progress: number;
	    scoreRaw: number;
	    notes: string;
	    repeat: number;
	    private: boolean;
	    startedYear: number;
	    startedMonth: number;
	    startedDay: number;
	    completedYear: number;
	    completedMonth: number;
	    completedDay: number;
	
	    static createFrom(source: any = {}) {
	        return new AnimeListEntryInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mediaId = source["mediaId"];
	        this.status = source["status"];
	        this.progress = source["progress"];
	        this.scoreRaw = source["scoreRaw"];
	        this.notes = source["notes"];
	        this.repeat = source["repeat"];
	        this.private = source["private"];
	        this.startedYear = source["startedYear"];
	        this.startedMonth = source["startedMonth"];
	        this.startedDay = source["startedDay"];
	        this.completedYear = source["completedYear"];
	        this.completedMonth = source["completedMonth"];
	        this.completedDay = source["completedDay"];
	    }
	}
	export class AnimeView {
	    id: number;
	    titleRomaji: string;
	    titleEnglish: string;
	    coverImage: string;
	    totalEpisodes: number;
	    status: string;
	    synopsis: string;
	    listStatus: string;
	
	    static createFrom(source: any = {}) {
	        return new AnimeView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.titleRomaji = source["titleRomaji"];
	        this.titleEnglish = source["titleEnglish"];
	        this.coverImage = source["coverImage"];
	        this.totalEpisodes = source["totalEpisodes"];
	        this.status = source["status"];
	        this.synopsis = source["synopsis"];
	        this.listStatus = source["listStatus"];
	    }
	}
	export class EpisodeView {
	    id: number;
	    anilistId: number;
	    episodeNumber: number;
	    filePath: string;
	    displayTitle: string;
	    status: string;
	    animeTitle: string;
	    coverImage: string;
	    bound: boolean;
	    progress: number;
	    totalEpisodes: number;
	    mediaStatus: string;
	    nextAiringEpisode: number;
	    resumePosition: number;
	    lastPlayedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new EpisodeView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.anilistId = source["anilistId"];
	        this.episodeNumber = source["episodeNumber"];
	        this.filePath = source["filePath"];
	        this.displayTitle = source["displayTitle"];
	        this.status = source["status"];
	        this.animeTitle = source["animeTitle"];
	        this.coverImage = source["coverImage"];
	        this.bound = source["bound"];
	        this.progress = source["progress"];
	        this.totalEpisodes = source["totalEpisodes"];
	        this.mediaStatus = source["mediaStatus"];
	        this.nextAiringEpisode = source["nextAiringEpisode"];
	        this.resumePosition = source["resumePosition"];
	        this.lastPlayedAt = source["lastPlayedAt"];
	    }
	}
	export class FuzzyDateView {
	    year: number;
	    month: number;
	    day: number;
	
	    static createFrom(source: any = {}) {
	        return new FuzzyDateView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.year = source["year"];
	        this.month = source["month"];
	        this.day = source["day"];
	    }
	}
	export class ImportResult {
	    episode: EpisodeView;
	    candidates: AnimeView[];
	    autoBound: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.episode = this.convertValues(source["episode"], EpisodeView);
	        this.candidates = this.convertValues(source["candidates"], AnimeView);
	        this.autoBound = source["autoBound"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class NyaaResultView {
	    title: string;
	    link: string;
	    magnet: string;
	    published: string;
	    size: string;
	    seeders: number;
	    leechers: number;
	    downloads: number;
	    trusted: boolean;
	    remake: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NyaaResultView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.link = source["link"];
	        this.magnet = source["magnet"];
	        this.published = source["published"];
	        this.size = source["size"];
	        this.seeders = source["seeders"];
	        this.leechers = source["leechers"];
	        this.downloads = source["downloads"];
	        this.trusted = source["trusted"];
	        this.remake = source["remake"];
	    }
	}
	export class RSSFeedItemView {
	    id: number;
	    feedId: number;
	    feedTitle: string;
	    title: string;
	    link: string;
	    magnet: string;
	    published: string;
	    isNew: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RSSFeedItemView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.feedId = source["feedId"];
	        this.feedTitle = source["feedTitle"];
	        this.title = source["title"];
	        this.link = source["link"];
	        this.magnet = source["magnet"];
	        this.published = source["published"];
	        this.isNew = source["isNew"];
	    }
	}
	export class RSSFeedView {
	    id: number;
	    url: string;
	    title: string;
	    enabled: boolean;
	    lastPolled: string;
	    newCount: number;
	
	    static createFrom(source: any = {}) {
	        return new RSSFeedView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.url = source["url"];
	        this.title = source["title"];
	        this.enabled = source["enabled"];
	        this.lastPolled = source["lastPolled"];
	        this.newCount = source["newCount"];
	    }
	}
	export class SettingsView {
	    mpvPath: string;
	    anime4kEnabled: boolean;
	    anime4kShadersReady: boolean;
	    downloadDir: string;
	    syncThreshold: number;
	    downloadRateLimit: number;
	    uploadRateLimit: number;
	    maxConcurrentDownloads: number;
	    seedRatio: number;
	    networkMode: string;
	    socks5Address: string;
	    httpProxyUrl: string;
	    updateChannel: string;
	    rssPollIntervalMinutes: number;
	    discordRpcEnabled: boolean;
	    downloadNotifications: boolean;
	    rssAutoDownload: boolean;
	    rssAutoDownloadLibraryOnly: boolean;
	    closeToTray: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SettingsView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mpvPath = source["mpvPath"];
	        this.anime4kEnabled = source["anime4kEnabled"];
	        this.anime4kShadersReady = source["anime4kShadersReady"];
	        this.downloadDir = source["downloadDir"];
	        this.syncThreshold = source["syncThreshold"];
	        this.downloadRateLimit = source["downloadRateLimit"];
	        this.uploadRateLimit = source["uploadRateLimit"];
	        this.maxConcurrentDownloads = source["maxConcurrentDownloads"];
	        this.seedRatio = source["seedRatio"];
	        this.networkMode = source["networkMode"];
	        this.socks5Address = source["socks5Address"];
	        this.httpProxyUrl = source["httpProxyUrl"];
	        this.updateChannel = source["updateChannel"];
	        this.rssPollIntervalMinutes = source["rssPollIntervalMinutes"];
	        this.discordRpcEnabled = source["discordRpcEnabled"];
	        this.downloadNotifications = source["downloadNotifications"];
	        this.rssAutoDownload = source["rssAutoDownload"];
	        this.rssAutoDownloadLibraryOnly = source["rssAutoDownloadLibraryOnly"];
	        this.closeToTray = source["closeToTray"];
	    }
	}
	export class StreamingEpisodeThumbnailView {
	    episodeNumber: number;
	    thumbnail: string;
	
	    static createFrom(source: any = {}) {
	        return new StreamingEpisodeThumbnailView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.episodeNumber = source["episodeNumber"];
	        this.thumbnail = source["thumbnail"];
	    }
	}
	export class UpdateInfo {
	    current: string;
	    latest: string;
	    available: boolean;
	    notes: string;
	    releaseUrl: string;
	    assetName: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.current = source["current"];
	        this.latest = source["latest"];
	        this.available = source["available"];
	        this.notes = source["notes"];
	        this.releaseUrl = source["releaseUrl"];
	        this.assetName = source["assetName"];
	    }
	}
	export class WatchingEntryView {
	    mediaId: number;
	    listStatus: string;
	    progress: number;
	    scoreRaw: number;
	    notes: string;
	    repeat: number;
	    private: boolean;
	    startedAt: FuzzyDateView;
	    completedAt: FuzzyDateView;
	    titleRomaji: string;
	    titleEnglish: string;
	    coverImage: string;
	    bannerImage: string;
	    totalEpisodes: number;
	    mediaStatus: string;
	    nextAiringEpisode: number;
	
	    static createFrom(source: any = {}) {
	        return new WatchingEntryView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mediaId = source["mediaId"];
	        this.listStatus = source["listStatus"];
	        this.progress = source["progress"];
	        this.scoreRaw = source["scoreRaw"];
	        this.notes = source["notes"];
	        this.repeat = source["repeat"];
	        this.private = source["private"];
	        this.startedAt = this.convertValues(source["startedAt"], FuzzyDateView);
	        this.completedAt = this.convertValues(source["completedAt"], FuzzyDateView);
	        this.titleRomaji = source["titleRomaji"];
	        this.titleEnglish = source["titleEnglish"];
	        this.coverImage = source["coverImage"];
	        this.bannerImage = source["bannerImage"];
	        this.totalEpisodes = source["totalEpisodes"];
	        this.mediaStatus = source["mediaStatus"];
	        this.nextAiringEpisode = source["nextAiringEpisode"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace torrentx {
	
	export class FileView {
	    path: string;
	    length: number;
	    bytesCompleted: number;
	    selected: boolean;
	    isVideo: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FileView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.length = source["length"];
	        this.bytesCompleted = source["bytesCompleted"];
	        this.selected = source["selected"];
	        this.isVideo = source["isVideo"];
	    }
	}
	export class ContentsView {
	    name: string;
	    bytesTotal: number;
	    files: FileView[];
	
	    static createFrom(source: any = {}) {
	        return new ContentsView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.bytesTotal = source["bytesTotal"];
	        this.files = this.convertValues(source["files"], FileView);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class JobView {
	    id: number;
	    name: string;
	    status: string;
	    bytesCompleted: number;
	    bytesTotal: number;
	    bytesUploaded: number;
	    percent: number;
	    uploadRatio: number;
	    speedBytesPerSecond: number;
	    uploadSpeedBytesPerSecond: number;
	    error: string;
	    source: string;
	    live: boolean;
	    files: FileView[];
	
	    static createFrom(source: any = {}) {
	        return new JobView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.status = source["status"];
	        this.bytesCompleted = source["bytesCompleted"];
	        this.bytesTotal = source["bytesTotal"];
	        this.bytesUploaded = source["bytesUploaded"];
	        this.percent = source["percent"];
	        this.uploadRatio = source["uploadRatio"];
	        this.speedBytesPerSecond = source["speedBytesPerSecond"];
	        this.uploadSpeedBytesPerSecond = source["uploadSpeedBytesPerSecond"];
	        this.error = source["error"];
	        this.source = source["source"];
	        this.live = source["live"];
	        this.files = this.convertValues(source["files"], FileView);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

