import {
  AddPlaylistSong,
  CreatePlaylist,
  DeletePlaylist,
  GetPlaylist,
  ListPlaylists,
  RemovePlaylistSong,
} from '../../../wailsjs/go/main/App';
import { music } from '../../../wailsjs/go/models';
import { favoriteSongKey, type FavoriteSong } from '@/lib/library/favoriteSong';
import {
  clearPlaylistDetailCache,
  deleteCachedPlaylist,
  getCachedPlaylist,
  getPlaylists,
  refreshPlaylistsFromBackend,
  setCachedPlaylist,
} from '@/stores/library/playlistsStore.svelte';
import { PLAYLISTS_UPDATED_EVENT } from '@/stores/library/playlistsStore.svelte';
import { EventsOn } from '../../../wailsjs/runtime/runtime';

export type UserPlaylist = {
  id: string;
  name: string;
  createdAt: string;
  songCount: number;
  songs: FavoriteSong[];
};

export function normalizePlaylist(raw: {
  id?: string;
  name?: string;
  createdAt?: string | { Time?: string };
  songs?: unknown[] | null;
}): UserPlaylist {
  let createdAt = '';
  if (typeof raw.createdAt === 'string') {
    createdAt = raw.createdAt;
  } else if (raw.createdAt && typeof raw.createdAt === 'object' && 'Time' in raw.createdAt) {
    createdAt = String(raw.createdAt.Time ?? '');
  }

  const songs = Array.isArray(raw.songs)
    ? raw.songs
        .map((item) => normalizeSong(item as Parameters<typeof normalizeSong>[0]))
        .filter((item) => item.title || item.id)
    : [];

  return {
    id: raw.id?.trim() ?? '',
    name: raw.name?.trim() ?? '',
    createdAt,
    songCount: songs.length,
    songs,
  };
}

function normalizeSong(raw: {
  id?: string;
  title?: string;
  artist?: string;
  album?: string;
  duration?: string;
  coverUrl?: string;
  sourceId?: string;
  platform?: string;
  metaJson?: string;
}): FavoriteSong {
  return {
    id: raw.id?.trim() ?? '',
    title: raw.title?.trim() ?? '',
    artist: raw.artist?.trim() ?? '',
    album: raw.album?.trim() ?? '',
    duration: raw.duration?.trim() ?? '',
    coverUrl: raw.coverUrl?.trim() ?? '',
    sourceId: raw.sourceId?.trim() ?? '',
    platform: raw.platform?.trim() ?? '',
    metaJson: raw.metaJson?.trim() ?? '',
  };
}

export function onPlaylistsChanged(listener: () => void): () => void {
  return EventsOn(PLAYLISTS_UPDATED_EVENT, () => listener());
}

/** 读取歌单列表；默认用内存状态，force 时从 Go 重新拉取 */
export async function fetchPlaylists(options?: { force?: boolean }): Promise<UserPlaylist[]> {
  if (!options?.force && getPlaylists().length > 0) {
    return [...getPlaylists()];
  }
  const items = await ListPlaylists();
  return items.map((item) => normalizePlaylist(item)).filter((item) => item.id && item.name);
}

export async function fetchPlaylist(
  id: string,
  options?: { force?: boolean },
): Promise<UserPlaylist | null> {
  const playlistId = id.trim();
  if (!playlistId) return null;

  if (!options?.force) {
    const cached = getCachedPlaylist(playlistId);
    if (cached) return cached;
  } else {
    deleteCachedPlaylist(playlistId);
  }

  try {
    const item = await GetPlaylist(playlistId);
    const playlist = normalizePlaylist(item);
    if (!playlist.id) return null;
    setCachedPlaylist(playlist);
    return playlist;
  } catch {
    return null;
  }
}

export async function createUserPlaylist(name: string): Promise<UserPlaylist> {
  const created = await CreatePlaylist(name.trim());
  return normalizePlaylist(created);
}

function toWailsFavoriteSong(song: FavoriteSong): music.FavoriteSong {
  return music.FavoriteSong.createFrom({
    id: song.id,
    title: song.title,
    artist: song.artist,
    album: song.album,
    duration: song.duration,
    coverUrl: song.coverUrl,
    sourceId: song.sourceId,
    platform: song.platform,
    metaJson: song.metaJson,
  });
}

export async function addSongToPlaylist(playlistId: string, song: FavoriteSong): Promise<void> {
  await AddPlaylistSong(playlistId, toWailsFavoriteSong(song));
  await refreshPlaylistsFromBackend();
}

export async function removeSongFromPlaylist(playlistId: string, song: FavoriteSong): Promise<void> {
  await RemovePlaylistSong(playlistId, toWailsFavoriteSong(song));
  await refreshPlaylistsFromBackend();
}

export async function deleteUserPlaylist(playlistId: string): Promise<void> {
  await DeletePlaylist(playlistId);
  clearPlaylistDetailCache();
}

export function playlistContainsSong(playlist: UserPlaylist, song: FavoriteSong): boolean {
  const key = favoriteSongKey(song);
  return playlist.songs.some((item) => favoriteSongKey(item) === key);
}

export function playlistCreateErrorMessage(err: unknown): string {
  const message = err instanceof Error ? err.message : String(err ?? '');
  if (message.includes('playlist name is required') || message.includes('歌单名称')) {
    return '请输入歌单名称';
  }
  if (message.includes('playlist name already exists') || message.includes('已存在')) {
    return '歌单名称已存在';
  }
  return message || '创建歌单失败';
}

export function playlistActionErrorMessage(err: unknown, fallback: string): string {
  const message = err instanceof Error ? err.message : String(err ?? '');
  if (message.includes('playlist not found')) {
    return '歌单不存在或已被删除';
  }
  if (message.includes('invalid song')) {
    return '无法添加该歌曲';
  }
  return message || fallback;
}
