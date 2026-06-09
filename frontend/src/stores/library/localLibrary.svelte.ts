import {
  fetchLocalSongCovers,
  GetLocalFolderSongsPage,
  GetLocalLibrarySnapshot,
  ScanLocalLibrary,
  localSongToTrackItem,
  type LocalSong,
} from '@/lib/library/localMusic';
import { error as toastError, success as toastSuccess } from '@/stores/ui/toast';
import type { TrackItem } from '@/lib/playback/track';
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import { music } from '../../../wailsjs/go/models';

export const LOCAL_ALL_TAB_ID = 'all';

export const LOCAL_LIBRARY_UPDATED_EVENT = 'local-library:updated';
export const LOCAL_LIBRARY_SCANNING_EVENT = 'local-library:scanning';

type LocalLibrarySnapshot = {
  folders: string[];
  folderAliases?: Record<string, string>;
  folderCounts?: Record<string, number>;
};

let lastMetaKey = '';
let syncInitialized = false;
let tabLoadToken = 0;
let coverLoadToken = 0;
let pendingCoverUpdates: Record<string, string> = {};
let coverFlushScheduled = false;
let localPageActive = false;
let activeTabId = LOCAL_ALL_TAB_ID;
let wasScanning = false;
let scanStart = 0;

const COVER_CHUNK = 40;
const COVER_FLUSH_PER_FRAME = 12;
const SONG_PAGE_SIZE = 1500;

function isTabLoaded(tabId: string): boolean {
  return tabId in localLibrary.loadedTabIds;
}

function applyCoverForPath(path: string, cover: string) {
  localLibrary.coverByPath[path] = cover;
  localLibrary.coverTickByPath[path] = (localLibrary.coverTickByPath[path] ?? 0) + 1;
}

function flushCoverUpdates() {
  coverFlushScheduled = false;
  const entries = Object.entries(pendingCoverUpdates);
  if (entries.length === 0) {
    return;
  }

  pendingCoverUpdates = {};
  const slice = entries.slice(0, COVER_FLUSH_PER_FRAME);
  for (const [path, cover] of slice) {
    applyCoverForPath(path, cover);
  }

  if (entries.length > COVER_FLUSH_PER_FRAME) {
    for (let i = COVER_FLUSH_PER_FRAME; i < entries.length; i += 1) {
      pendingCoverUpdates[entries[i][0]] = entries[i][1];
    }
    scheduleCoverFlush();
  }
}

function scheduleCoverFlush() {
  if (coverFlushScheduled) {
    return;
  }
  coverFlushScheduled = true;
  requestAnimationFrame(flushCoverUpdates);
}

function collectUniqueLibraryPaths(): string[] {
  const paths = new Set<string>();
  for (const tabId of Object.keys(localLibrary.loadedTabIds)) {
    for (const track of localLibrary.tracksByTab[tabId] ?? []) {
      const path = String(track.listKey ?? track.id).trim();
      if (path) {
        paths.add(path);
      }
    }
  }
  return [...paths];
}

async function loadCoversForPaths(paths: string[]): Promise<void> {
  const pending = paths.filter(
    (path) => path && !localLibrary.coverByPath[path] && !pendingCoverUpdates[path],
  );
  if (pending.length === 0) {
    return;
  }

  const token = coverLoadToken;
  const batch = await fetchLocalSongCovers(pending);
  if (token !== coverLoadToken) {
    return;
  }

  let added = 0;
  for (const [path, key] of Object.entries(batch.paths)) {
    const cover = batch.covers[key];
    if (cover && !localLibrary.coverByPath[path] && !pendingCoverUpdates[path]) {
      pendingCoverUpdates[path] = cover;
      added += 1;
    }
  }
  if (added > 0) {
    scheduleCoverFlush();
  }
}

function yieldToMain(): Promise<void> {
  return new Promise((resolve) => {
    setTimeout(resolve, 0);
  });
}

function collectPathsForTab(tabId: string): string[] {
  const tracks = localLibrary.tracksByTab[tabId] ?? [];
  const paths: string[] = [];
  for (const track of tracks) {
    const path = String(track.listKey ?? track.id).trim();
    if (path) {
      paths.push(path);
    }
  }
  return paths;
}

function orderedCoverPreloadPaths(): string[] {
  const activePaths = collectPathsForTab(activeTabId);
  const seen = new Set(activePaths);
  const ordered = [...activePaths];
  for (const path of collectUniqueLibraryPaths()) {
    if (!seen.has(path)) {
      seen.add(path);
      ordered.push(path);
    }
  }
  return ordered;
}

/** 曲目索引就绪后，在空闲时把封面批量写入 store（内存缓存，切 Tab 不重拉） */
function scheduleCoverPreloadAll() {
  if (!localPageActive) {
    return;
  }
  const paths = orderedCoverPreloadPaths();
  if (paths.length === 0) {
    return;
  }

  const token = ++coverLoadToken;
  const run = async () => {
    for (let i = 0; i < paths.length; i += COVER_CHUNK) {
      if (coverLoadToken !== token) {
        return;
      }
      if (!localPageActive) {
        return;
      }
      await loadCoversForPaths(paths.slice(i, i + COVER_CHUNK));
      await yieldToMain();
    }
  };

  const start = () => void run();
  setTimeout(() => {
    if (!localPageActive) {
      return;
    }
    if (typeof requestIdleCallback !== 'undefined') {
      requestIdleCallback(start, { timeout: 4000 });
    } else {
      start();
    }
  }, 600);
}

/** 进入本地页或 Tab 曲目就绪时：先拉当前 Tab 可见封面，再后台预加载其余 */
function ensureCoversForCurrentView() {
  if (!localPageActive || !isTabLoaded(activeTabId)) {
    return;
  }
  const paths = collectPathsForTab(activeTabId).slice(0, COVER_CHUNK * 2);
  if (paths.length > 0) {
    void loadCoversForPaths(paths);
  }
  scheduleCoverPreloadAll();
}

function clearCoverCache() {
  coverLoadToken += 1;
  pendingCoverUpdates = {};
  coverFlushScheduled = false;
  localLibrary.coverByPath = {};
  localLibrary.coverTickByPath = {};
}

/** 写入本地封面缓存并 bump 对应 path 的 tick */
export function applyLocalCoverForPath(path: string, cover: string) {
  applyCoverForPath(path, cover);
}

/** 读取已缓存的本地封面 URL */
export function localCoverUrl(path: string): string {
  return localLibrary.coverByPath[path]?.trim() ?? '';
}

export const localLibrary = $state({
  folders: [] as string[],
  folderAliases: {} as Record<string, string>,
  folderCounts: {} as Record<string, number>,
  tracksByTab: {} as Record<string, TrackItem[]>,
  loadedTabIds: {} as Record<string, true>,
  loadingTabId: null as string | null,
  songById: new Map<string, LocalSong>(),
  coverByPath: {} as Record<string, string>,
  coverTickByPath: {} as Record<string, number>,
  revision: 0,
  loading: false,
  scanning: false,
  loaded: false,
});

export function setLocalPageActive(active: boolean): void {
  localPageActive = active;
  if (!active) {
    return;
  }
  if (localLibrary.loaded && !isTabLoaded(activeTabId)) {
    void loadTabTracks(activeTabId);
    return;
  }
  ensureCoversForCurrentView();
}

export function setLocalActiveFolderTab(tabId: string): void {
  activeTabId = tabId;
  if (!localPageActive) {
    return;
  }
  if (!isTabLoaded(tabId)) {
    void loadTabTracks(tabId);
    return;
  }
  const paths = collectPathsForTab(tabId).slice(0, COVER_CHUNK * 2);
  if (paths.length > 0) {
    void loadCoversForPaths(paths);
  }
}

function snapshotMetaKey(snapshot: LocalLibrarySnapshot): string {
  const foldersKey = snapshot.folders.join('\0');
  const counts = snapshot.folderCounts ?? {};
  const countKeys = Object.keys(counts).sort();
  let countsPart = '';
  for (const key of countKeys) {
    countsPart += `${key}\x01${counts[key]}\x02`;
  }
  return `${foldersKey}\x00${countsPart}`;
}

function buildTracksFromSongs(songs: LocalSong[], songById: Map<string, LocalSong>): TrackItem[] {
  const tracks: TrackItem[] = [];
  const trackByPath = new Map<string, TrackItem>();

  for (const song of songs) {
    songById.set(String(song.id), song);
    let item = trackByPath.get(song.filePath);
    if (!item) {
      item = localSongToTrackItem(song);
      trackByPath.set(song.filePath, item);
    }
    tracks.push(item);
  }

  return tracks;
}

function appendTracksFromSongs(
  songs: LocalSong[],
  tracks: TrackItem[],
  songById: Map<string, LocalSong>,
): TrackItem[] {
  if (songs.length === 0) {
    return tracks;
  }
  const nextTracks = tracks.slice();
  const trackByPath = new Map<string, TrackItem>();
  for (const track of nextTracks) {
    const path = String(track.listKey ?? track.id).trim();
    if (path) {
      trackByPath.set(path, track);
    }
  }

  for (const song of songs) {
    songById.set(String(song.id), song);
    let item = trackByPath.get(song.filePath);
    if (!item) {
      item = localSongToTrackItem(song);
      trackByPath.set(song.filePath, item);
    }
    nextTracks.push(item);
  }

  return nextTracks;
}

function applyTabTracks(tabId: string, songs: LocalSong[]) {
  const songById = new Map(localLibrary.songById);
  localLibrary.tracksByTab[tabId] = buildTracksFromSongs(songs ?? [], songById);
  localLibrary.songById = songById;
  localLibrary.loadedTabIds[tabId] = true;
  localLibrary.revision += 1;
}

function invalidateTracksIndex() {
  tabLoadToken += 1;
  localLibrary.loadingTabId = null;
  localLibrary.tracksByTab = {};
  localLibrary.loadedTabIds = {};
  localLibrary.songById = new Map();
  localLibrary.revision += 1;
}

function normalizeSnapshot(raw: music.LocalLibrarySnapshot | LocalLibrarySnapshot): LocalLibrarySnapshot {
  return {
    folders: raw.folders ?? [],
    folderAliases: raw.folderAliases ?? {},
    folderCounts: raw.folderCounts ?? {},
  };
}

function applySnapshot(snapshot: LocalLibrarySnapshot) {
  const metaKey = snapshotMetaKey(snapshot);
  if (metaKey !== lastMetaKey) {
    if (lastMetaKey !== '') {
      clearCoverCache();
      invalidateTracksIndex();
    }
    lastMetaKey = metaKey;
  }

  localLibrary.folders = snapshot.folders;
  localLibrary.folderAliases = snapshot.folderAliases ?? {};
  localLibrary.folderCounts = snapshot.folderCounts ?? {};
  localLibrary.loaded = true;

  if (localPageActive) {
    void loadTabTracks(activeTabId);
  }
}

function tabCountsHaveSongs(tabId: string): boolean {
  return (localLibrary.folderCounts[tabId] ?? 0) > 0;
}

/** 按 Tab 懒加载曲目（分页 IPC，避免大库一次性传输阻塞） */
export async function loadTabTracks(tabId: string): Promise<void> {
  if (isTabLoaded(tabId)) {
    ensureCoversForCurrentView();
    return;
  }

  const token = ++tabLoadToken;
  localLibrary.loadingTabId = tabId;

  try {
    let offset = 0;
    let total = 0;
    let tracks: TrackItem[] = [];
    const songById = new Map(localLibrary.songById);

    while (true) {
      const page = (await GetLocalFolderSongsPage(tabId, offset, SONG_PAGE_SIZE)) as {
        songs?: LocalSong[];
        total?: number;
        offset?: number;
      };
      if (tabLoadToken !== token) {
        return;
      }

      const chunk = page.songs ?? [];
      total = page.total ?? chunk.length;
      tracks = offset === 0 ? buildTracksFromSongs(chunk, songById) : appendTracksFromSongs(chunk, tracks, songById);
      localLibrary.tracksByTab[tabId] = tracks;
      localLibrary.songById = songById;
      localLibrary.revision += 1;

      if (offset === 0) {
        localLibrary.loadedTabIds[tabId] = true;
        ensureCoversForCurrentView();
      }

      offset += chunk.length;
      if (chunk.length === 0 || offset >= total) {
        break;
      }
      await yieldToMain();
    }
  } catch (err) {
    if (tabLoadToken === token) {
      applyTabTracks(tabId, []);
      const message = err instanceof Error ? err.message : String(err);
      toastError(message ? `加载本地歌曲失败：${message}` : '加载本地歌曲失败');
    }
  } finally {
    if (tabLoadToken === token) {
      localLibrary.loadingTabId = null;
    }
  }
}

/** @deprecated 使用 loadTabTracks；保留兼容旧调用 */
export async function loadTracksIndex(): Promise<void> {
  await loadTabTracks(activeTabId);
}

export function isLocalTabLoaded(tabId: string): boolean {
  return isTabLoaded(tabId);
}

export function isLocalTabLoading(tabId: string): boolean {
  return localLibrary.loadingTabId === tabId;
}

export function isLocalTabFullyLoaded(tabId: string): boolean {
  if (!isTabLoaded(tabId)) {
    return false;
  }
  const expected = localLibrary.folderCounts[tabId] ?? localLibrary.tracksByTab[tabId]?.length ?? 0;
  const loaded = localLibrary.tracksByTab[tabId]?.length ?? 0;
  return loaded >= expected;
}

export function localTabHasTracks(tabId: string): boolean {
  if (isTabLoaded(tabId)) {
    return (localLibrary.tracksByTab[tabId]?.length ?? 0) > 0;
  }
  return tabCountsHaveSongs(tabId);
}

export function initLocalLibrarySync(): () => void {
  if (syncInitialized) {
    return () => {};
  }
  syncInitialized = true;

  const offUpdated = EventsOn(LOCAL_LIBRARY_UPDATED_EVENT, (payload: music.LocalLibrarySnapshot) => {
    applySnapshot(normalizeSnapshot(payload));
  });
  const offScanning = EventsOn(LOCAL_LIBRARY_SCANNING_EVENT, (scanning: boolean) => {
    const nextScanning = Boolean(scanning);
    if (!wasScanning && nextScanning) {
      scanStart = Date.now();
    }
    if (wasScanning && !nextScanning) {
      if (scanStart > 0) {
        const seconds = (Date.now() - scanStart) / 1000;
        toastSuccess(`本地音乐扫描完成，用时 ${seconds.toFixed(1)} 秒`);
        scanStart = 0;
      }
      if (localLibrary.loaded) {
        invalidateTracksIndex();
        if (localPageActive) {
          void loadTabTracks(activeTabId);
        }
      }
    }
    wasScanning = nextScanning;
    localLibrary.scanning = nextScanning;
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

export async function scanLocalLibrary(): Promise<void> {
  await ScanLocalLibrary();
}
