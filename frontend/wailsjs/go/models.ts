export namespace music {
	
	export class LyricInfo {
	    lyric: string;
	    tlyric?: string;
	    rlyric?: string;
	    lxlyric?: string;
	
	    static createFrom(source: any = {}) {
	        return new LyricInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lyric = source["lyric"];
	        this.tlyric = source["tlyric"];
	        this.rlyric = source["rlyric"];
	        this.lxlyric = source["lxlyric"];
	    }
	}
	export class PlatformInfo {
	    key: string;
	    name: string;
	    actions: string[];
	    qualities: string[];
	
	    static createFrom(source: any = {}) {
	        return new PlatformInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.name = source["name"];
	        this.actions = source["actions"];
	        this.qualities = source["qualities"];
	    }
	}
	export class SongItem {
	    id: string;
	    name: string;
	    singer: string;
	    album: string;
	    albumId?: string;
	    source: string;
	    interval?: string;
	    img?: string;
	    songmid: string;
	    hash?: string;
	    metaJson: string;
	
	    static createFrom(source: any = {}) {
	        return new SongItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.singer = source["singer"];
	        this.album = source["album"];
	        this.albumId = source["albumId"];
	        this.source = source["source"];
	        this.interval = source["interval"];
	        this.img = source["img"];
	        this.songmid = source["songmid"];
	        this.hash = source["hash"];
	        this.metaJson = source["metaJson"];
	    }
	}
	export class SearchResult {
	    list: SongItem[];
	    total: number;
	    page: number;
	    limit: number;
	    source: string;
	
	    static createFrom(source: any = {}) {
	        return new SearchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.list = this.convertValues(source["list"], SongItem);
	        this.total = source["total"];
	        this.page = source["page"];
	        this.limit = source["limit"];
	        this.source = source["source"];
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
	
	export class SourceInfo {
	    id: string;
	    name: string;
	    description: string;
	    version: string;
	    author: string;
	    homepage: string;
	    filename: string;
	    enabled: boolean;
	    platforms: PlatformInfo[];
	    // Go type: time
	    importedAt: any;
	    status: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new SourceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.version = source["version"];
	        this.author = source["author"];
	        this.homepage = source["homepage"];
	        this.filename = source["filename"];
	        this.enabled = source["enabled"];
	        this.platforms = this.convertValues(source["platforms"], PlatformInfo);
	        this.importedAt = this.convertValues(source["importedAt"], null);
	        this.status = source["status"];
	        this.error = source["error"];
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

