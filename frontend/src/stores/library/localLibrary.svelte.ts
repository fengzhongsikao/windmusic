import {
  fetchLocalSongCovers,
  GetLocalLibrarySnapshot,
  GetLocalLibraryTracksIndex,
  ScanLocalLibrary,
  localSongToTrackItem,
  type LocalSong,
} from '@/lib/library/localMusic';
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
let tracksIndexToken = 0;
let coverLoadToken = 0;
let pendingCoverUpdates: Record<string, string> = {};
let coverFlushScheduled = false;
let localPageActive = false;
let activeTabId = LOCAL_ALL_TAB_ID;

const COVER_CHUNK = 40;
const COVER_FLUSH_PER_FRAME = 12;

function flushCoverUpdates() {
  coverFlushScheduled = false;
  const entries = Object.entries(pendingCoverUpdates);
  if (entries.length === 0) {
    return;
  }

  pendingCoverUpdates = {};
  const slice = entries.slice(0, COVER_FLUSH_PER_FRAME);
  for (const [path, cover] of slice) {
    localLibrary.coverByPath[path] = cover;
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
  for (const tracks of Object.values(localLibrary.tracksByTab)) {
    for (const track of tracks) {
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
  // 曲目列表先稳定，再在空闲时灌封面，避免和 Tab 点击抢主线程
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

/** 进入本地页或索引刚就绪时：先拉当前 Tab 可见封面，再后台预加载其余 */
function ensureCoversForCurrentView() {
  if (!localPageActive || !localLibrary.tracksIndexReady) {
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
}

export const localLibrary = $state({
  folders: [] as string[],
  folderAliases: {} as Record<string, string>,
  folderCounts: {} as Record<string, number>,
  tracksByTab: {} as Record<string, TrackItem[]>,
  songById: new Map<string, LocalSong>(),
  tracksIndexReady: false,
  tracksIndexLoading: false,
  coverByPath: {} as Record<string, string>,
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
  if (localLibrary.loaded && !localLibrary.tracksIndexReady) {
    void loadTracksIndex();
    return;
  }
  ensureCoversForCurrentView();
}

export function setLocalActiveFolderTab(tabId: string): void {
  activeTabId = tabId;
  if (!localPageActive || !localLibrary.tracksIndexReady) {
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

function buildTracksFromSongs(songs: LocalSong[]): TrackItem[] {
  const tracks: TrackItem[] = [];
  const trackByPath = new Map<string, TrackItem>();

  for (const song of songs) {
    localLibrary.songById.set(String(song.id), song);
    let item = trackByPath.get(song.filePath);
    if (!item) {
      item = localSongToTrackItem(song);
      trackByPath.set(song.filePath, item);
    }
    tracks.push(item);
  }

  return tracks;
}

function applyTracksIndex(index: Record<string, LocalSong[]>) {
  const tracksByTab: Record<string, TrackItem[]> = {};
  localLibrary.songById = new Map();

  for (const [tabId, songs] of Object.entries(index)) {
    tracksByTab[tabId] = buildTracksFromSongs(songs ?? []);
  }

  localLibrary.tracksByTab = tracksByTab;
  localLibrary.tracksIndexReady = true;
  localLibrary.revision += 1;
}

function invalidateTracksIndex() {
  tracksIndexToken += 1;
  localLibrary.tracksIndexLoading = false;
  localLibrary.tracksByTab = {};
  localLibrary.songById = new Map();
  localLibrary.tracksIndexReady = false;
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
    void loadTracksIndex();
  }
}

function tracksIndexHasSongs(): boolean {
  return Object.values(localLibrary.tracksByTab).some((tracks) => tracks.length > 0);
}

function snapshotCountsHaveSongs(): boolean {
  return Object.values(localLibrary.folderCounts).some((count) => count > 0);
}

/** 一次 IPC 拉取后端按文件夹分好的曲目索引（与扫描结构一致） */
export async function loadTracksIndex(): Promise<void> {
  if (
    localLibrary.tracksIndexReady &&
    (tracksIndexHasSongs() || !snapshotCountsHaveSongs())
  ) {
    ensureCoversForCurrentView();
    return;
  }

  const token = ++tracksIndexToken;
  localLibrary.tracksIndexLoading = true;

  try {
    const index = (await GetLocalLibraryTracksIndex()) as Record<string, LocalSong[]>;
    if (tracksIndexToken !== token) {
      return;
    }
    applyTracksIndex(index ?? {});
    ensureCoversForCurrentView();
  } catch {
    if (tracksIndexToken === token) {
      invalidateTracksIndex();
    }
  } finally {
    if (tracksIndexToken === token) {
      localLibrary.tracksIndexLoading = false;
    }
  }
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

export async function scanLocalLibrary(): Promise<void> {
  await ScanLocalLibrary();
}
