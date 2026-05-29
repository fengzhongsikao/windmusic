import {
  fetchLocalSongCovers,
  GetLocalLibrarySnapshot,
  ScanLocalLibrary,
  localSongToTrackItem,
  songInFolder,
  type LocalSong,
} from '@/lib/localMusic';
import type { TrackItem } from '@/lib/track';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { music } from '../../wailsjs/go/models';

export const LOCAL_ALL_TAB_ID = 'all';

export const LOCAL_LIBRARY_UPDATED_EVENT = 'local-library:updated';
export const LOCAL_LIBRARY_SCANNING_EVENT = 'local-library:scanning';

type LocalLibrarySnapshot = {
  folders: string[];
  songs: LocalSong[];
};

let coverLoadToken = 0;
let coverPreloadStarted = false;
let lastSongPathsKey = '';
let syncInitialized = false;

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

function normalizeSnapshot(raw: music.LocalLibrarySnapshot | LocalLibrarySnapshot): LocalLibrarySnapshot {
  return {
    folders: raw.folders ?? [],
    songs: (raw.songs ?? []) as LocalSong[],
  };
}

function applySnapshot(snapshot: LocalLibrarySnapshot) {
  const pathsKey = snapshot.songs
    .map((song) => song.filePath)
    .sort()
    .join('\0');
  if (pathsKey !== lastSongPathsKey) {
    if (lastSongPathsKey !== '') {
      clearLocalCoverCache();
    }
    lastSongPathsKey = pathsKey;
    scheduleCoverPreload(snapshot.songs.map((song) => song.filePath));
  }

  localLibrary.folders = snapshot.folders;
  localLibrary.songs = snapshot.songs;
  rebuildLibraryIndex(snapshot.songs, snapshot.folders);
  localLibrary.loaded = true;
}

/** 订阅后端推送；启动时拉一次快照，避免错过 startup 事件 */
export function initLocalLibrarySync(): () => void {
  if (syncInitialized) {
    return () => {};
  }
  syncInitialized = true;

  const offUpdated = EventsOn(LOCAL_LIBRARY_UPDATED_EVENT, (payload: music.LocalLibrarySnapshot) => {
    applySnapshot(normalizeSnapshot(payload));
  });
  const offScanning = EventsOn(LOCAL_LIBRARY_SCANNING_EVENT, (scanning: boolean) => {
    localLibrary.scanning = Boolean(scanning);
  });

  localLibrary.loading = true;
  void GetLocalLibrarySnapshot()
    .then((snapshot) => {
      applySnapshot(normalizeSnapshot(snapshot));
    })
    .catch(() => {})
    .finally(() => {
      localLibrary.loading = false;
    });

  return () => {
    offUpdated();
    offScanning();
    syncInitialized = false;
  };
}

/** 手动全盘扫描（后端扫描完成后通过事件更新状态） */
export async function scanLocalLibrary(): Promise<void> {
  await ScanLocalLibrary();
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
