import { ListPlaylists } from '../../../wailsjs/go/main/App';
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import { music } from '../../../wailsjs/go/models';
import { normalizePlaylist, type UserPlaylist } from '@/lib/playlists';

export const PLAYLISTS_UPDATED_EVENT = 'playlists:updated';

export const playlistsState = $state({
  items: [] as UserPlaylist[],
  loaded: false,
});

const playlistDetailCache = new Map<string, UserPlaylist>();

let syncInitialized = false;

function applyPlaylists(raw: music.UserPlaylist[]) {
  playlistsState.items = raw
    .map((item) => normalizePlaylist(item))
    .filter((item) => item.id && item.name);
  for (const item of playlistsState.items) {
    playlistDetailCache.set(item.id, item);
  }
  playlistsState.loaded = true;
}

export function initPlaylistsSync(): () => void {
  if (syncInitialized) {
    return () => {};
  }
  syncInitialized = true;

  const offUpdated = EventsOn(PLAYLISTS_UPDATED_EVENT, (payload: music.UserPlaylist[]) => {
    playlistDetailCache.clear();
    applyPlaylists(payload ?? []);
  });

  void ListPlaylists()
    .then((items) => applyPlaylists(items ?? []))
    .catch(() => {});

  return () => {
    offUpdated();
    syncInitialized = false;
  };
}

export function getPlaylists(): UserPlaylist[] {
  return playlistsState.items;
}

export function getCachedPlaylist(id: string): UserPlaylist | undefined {
  return playlistDetailCache.get(id) ?? playlistsState.items.find((item) => item.id === id);
}

export function setCachedPlaylist(playlist: UserPlaylist): void {
  playlistDetailCache.set(playlist.id, playlist);
}

export function deleteCachedPlaylist(id: string): void {
  playlistDetailCache.delete(id.trim());
}

export function clearPlaylistDetailCache(): void {
  playlistDetailCache.clear();
}

/** 从 Go 重新拉取歌单列表（添加歌曲后或进入歌单页时调用） */
export async function refreshPlaylistsFromBackend(): Promise<void> {
  try {
    const items = await ListPlaylists();
    playlistDetailCache.clear();
    applyPlaylists(items ?? []);
  } catch {
    // 保留现有内存状态
  }
}
