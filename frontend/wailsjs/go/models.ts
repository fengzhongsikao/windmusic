export namespace music {
	
	export class FavoriteSong {
	    id: string;
	    title: string;
	    artist: string;
	    album?: string;
	    duration?: string;
	    coverUrl?: string;
	    sourceId?: string;
	    platform?: string;
	    metaJson?: string;
	
	    static createFrom(source: any = {}) {
	        return new FavoriteSong(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.artist = source["artist"];
	        this.album = source["album"];
	        this.duration = source["duration"];
	        this.coverUrl = source["coverUrl"];
	        this.sourceId = source["sourceId"];
	        this.platform = source["platform"];
	        this.metaJson = source["metaJson"];
	    }
	}
	export class LocalCoverBatch {
	    covers: Record<string, string>;
	    paths: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new LocalCoverBatch(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.covers = source["covers"];
	        this.paths = source["paths"];
	    }
	}
	export class LocalSong {
	    id: string;
	    title: string;
	    artist: string;
	    album?: string;
	    duration?: string;
	    filePath: string;
	    format: string;
	    size: string;
	    coverData?: string;
	    lyric?: string;
	
	    static createFrom(source: any = {}) {
	        return new LocalSong(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.artist = source["artist"];
	        this.album = source["album"];
	        this.duration = source["duration"];
	        this.filePath = source["filePath"];
	        this.format = source["format"];
	        this.size = source["size"];
	        this.coverData = source["coverData"];
	        this.lyric = source["lyric"];
	    }
	}
	export class LocalSongExtras {
	    coverData?: string;
	    lyric?: string;
	
	    static createFrom(source: any = {}) {
	        return new LocalSongExtras(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.coverData = source["coverData"];
	        this.lyric = source["lyric"];
	    }
	}
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
	export class RecentSong {
	    id: string;
	    title: string;
	    artist: string;
	    album?: string;
	    duration?: string;
	    coverUrl?: string;
	    sourceId?: string;
	    platform?: string;
	    metaJson?: string;
	    // Go type: time
	    playedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new RecentSong(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.artist = source["artist"];
	        this.album = source["album"];
	        this.duration = source["duration"];
	        this.coverUrl = source["coverUrl"];
	        this.sourceId = source["sourceId"];
	        this.platform = source["platform"];
	        this.metaJson = source["metaJson"];
	        this.playedAt = this.convertValues(source["playedAt"], null);
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
	
	export class UserPlaylist {
	    id: string;
	    name: string;
	    // Go type: time
	    createdAt: any;
	    songs?: FavoriteSong[];
	
	    static createFrom(source: any = {}) {
	        return new UserPlaylist(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.songs = this.convertValues(source["songs"], FavoriteSong);
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

