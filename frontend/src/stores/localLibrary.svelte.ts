import {
  fetchLocalSongCovers,
  ListLocalFolders,
  ListLocalLibrary,
  localSongToTrackItem,
  ScanLocalLibrary,
  songInFolder,
  type LocalSong,
} from '@/lib/localMusic';
import type { TrackItem } from '@/lib/track';

export const LOCAL_ALL_TAB_ID = 'all';

type LocalLibrarySnapshot = {
  folders: string[];
  songs: LocalSong[];
};

let sessionCache: LocalLibrarySnapshot | null = null;
let coverLoadToken = 0;
let coverPreloadStarted = false;

export const localLibrary = $state({
  folders: [] as string[],
  songs: [] as LocalSong[],
  folderCounts: {} as Record<string, number>,
  tracksByTab: {} as Record<string, TrackItem[]>,
  coverByPath: {} as Record<string, string>,
  loading: false,
  scanning: false,
  loaded: false,
});

function rebuildLibraryIndex(songs: LocalSong[], folders: string[]) {
  const folderCounts: Record<string, number> = { [LOCAL_ALL_TAB_ID]: songs.length };
  const tracksByTab: Record<string, TrackItem[]> = {
    [LOCAL_ALL_TAB_ID]: songs.map((song) => localSongToTrackItem(song)),
  };

  for (const folder of folders) {
    const folderSongs: LocalSong[] = [];
    for (const song of songs) {
      if (songInFolder(song.filePath, folder)) {
        folderSongs.push(song);
      }
    }
    folderCounts[folder] = folderSongs.length;
    tracksByTab[folder] = folderSongs.map((song) => localSongToTrackItem(song));
  }

  localLibrary.folderCounts = folderCounts;
  localLibrary.tracksByTab = tracksByTab;
}

export function clearLocalCoverCache() {
  coverLoadToken += 1;
  coverPreloadStarted = false;
  localLibrary.coverByPath = {};
}

function applySnapshot(snapshot: LocalLibrarySnapshot) {
  localLibrary.folders = snapshot.folders;
  localLibrary.songs = snapshot.songs;
  rebuildLibraryIndex(snapshot.songs, snapshot.folders);
  sessionCache = snapshot;
  localLibrary.loaded = true;
  scheduleCoverPreload(snapshot.songs.map((song) => song.filePath));
}

/** 从磁盘缓存快速加载，不遍历文件夹 */
export async function loadLocalLibraryFromCache(): Promise<LocalLibrarySnapshot> {
  if (sessionCache) {
    applySnapshot(sessionCache);
    return sessionCache;
  }

  localLibrary.loading = true;
  try {
    const [folders, songs] = await Promise.all([ListLocalFolders(), ListLocalLibrary()]);
    const snapshot = { folders: folders ?? [], songs: songs ?? [] };
    applySnapshot(snapshot);
    return snapshot;
  } finally {
    localLibrary.loading = false;
  }
}

/** 全盘扫描（添加/移除文件夹、手动刷新时使用） */
export async function scanLocalLibrary(): Promise<LocalLibrarySnapshot> {
  localLibrary.scanning = true;
  try {
    const [folders, songs] = await Promise.all([ListLocalFolders(), ScanLocalLibrary()]);
    const snapshot = { folders: folders ?? [], songs: songs ?? [] };
    applySnapshot(snapshot);
    return snapshot;
  } finally {
    localLibrary.scanning = false;
  }
}

/** 进入页面：优先读缓存；仅有文件夹但尚无曲目时再后台扫描 */
export async function ensureLocalLibrary(): Promise<void> {
  const snapshot = await loadLocalLibraryFromCache();
  if (snapshot.folders.length > 0 && snapshot.songs.length === 0) {
    await scanLocalLibrary();
  }
}

export function invalidateLocalLibrarySession() {
  sessionCache = null;
  localLibrary.loaded = false;
  clearLocalCoverCache();
}

/** 后台批量加载封面（只读 extras 缓存，不扫盘） */
export async function loadLocalCoversForPaths(paths: string[]): Promise<void> {
  const pending = paths.filter((path) => path && !localLibrary.coverByPath[path]);
  if (pending.length === 0) {
    return;
  }

  const token = coverLoadToken;
  const batch = await fetchLocalSongCovers(pending);
  if (token !== coverLoadToken) {
    return;
  }

  const next = { ...localLibrary.coverByPath };
  let added = 0;
  for (const [path, key] of Object.entries(batch.paths)) {
    const cover = batch.covers[key];
    if (cover && !next[path]) {
      next[path] = cover;
      added += 1;
    }
  }
  if (added > 0) {
    localLibrary.coverByPath = next;
  }
}

const COVER_PRELOAD_CHUNK = 120;

/** 库加载后空闲时分批预取全部封面，避免切换 Tab 时再触发加载 */
export function scheduleCoverPreload(paths: string[]): void {
  if (coverPreloadStarted) {
    return;
  }
  const unique = [...new Set(paths.map((path) => path.trim()).filter(Boolean))];
  if (unique.length === 0) {
    return;
  }
  coverPreloadStarted = true;
  const tokenAtStart = coverLoadToken;

  const run = async () => {
    for (let i = 0; i < unique.length; i += COVER_PRELOAD_CHUNK) {
      if (coverLoadToken !== tokenAtStart) {
        return;
      }
      await loadLocalCoversForPaths(unique.slice(i, i + COVER_PRELOAD_CHUNK));
      await yieldToMain();
    }
  };

  const start = () => void run();
  if (typeof requestIdleCallback !== 'undefined') {
    requestIdleCallback(start, { timeout: 2000 });
  } else {
    setTimeout(start, 50);
  }
}

function yieldToMain(): Promise<void> {
  return new Promise((resolve) => {
    setTimeout(resolve, 0);
  });
}

/** 应用启动时后台预加载，进入本地音乐页时通常已就绪 */
export function preloadLocalLibrary(): void {
  if (sessionCache || localLibrary.loading || localLibrary.scanning) {
    return;
  }
  void loadLocalLibraryFromCache().then((snapshot) => {
    if (snapshot.folders.length > 0 && snapshot.songs.length === 0) {
      void scanLocalLibrary();
    }
  });
}
