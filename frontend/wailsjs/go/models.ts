export namespace main {
	
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
	export class AnimeView {
	    id: number;
	    titleRomaji: string;
	    titleEnglish: string;
	    coverImage: string;
	    totalEpisodes: number;
	    status: string;
	    synopsis: string;
	
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
	export class SettingsView {
	    mpvPath: string;
	    downloadDir: string;
	    syncThreshold: number;
	    anilistClientId: string;
	
	    static createFrom(source: any = {}) {
	        return new SettingsView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mpvPath = source["mpvPath"];
	        this.downloadDir = source["downloadDir"];
	        this.syncThreshold = source["syncThreshold"];
	        this.anilistClientId = source["anilistClientId"];
	    }
	}

}

export namespace torrentx {
	
	export class JobView {
	    id: number;
	    name: string;
	    status: string;
	    bytesCompleted: number;
	    bytesTotal: number;
	    percent: number;
	    error: string;
	    source: string;
	
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
	        this.percent = source["percent"];
	        this.error = source["error"];
	        this.source = source["source"];
	    }
	}

}

