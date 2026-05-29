import {
  AddPlaylistSong,
  CreatePlaylist,
  DeletePlaylist,
  GetPlaylist,
  ListPlaylists,
  RemovePlaylistSong,
} from '../../wailsjs/go/main/App';
import { music } from '../../wailsjs/go/models';
import { favoriteSongKey, type FavoriteSong } from '@/lib/wailsPlayer';

export type UserPlaylist = {
  id: string;
  name: string;
  createdAt: string;
  songCount: number;
  songs: FavoriteSong[];
};

const PLAYLISTS_CHANGED_EVENT = 'playlists:changed';

let playlistsCache: UserPlaylist[] | null = null;
const playlistDetailCache = new Map<string, UserPlaylist>();

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

function normalizePlaylist(raw: {
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

function notifyPlaylistsChanged() {
  playlistsCache = null;
  playlistDetailCache.clear();
  window.dispatchEvent(new CustomEvent(PLAYLISTS_CHANGED_EVENT));
}

export function onPlaylistsChanged(listener: () => void): () => void {
  const handler = () => listener();
  window.addEventListener(PLAYLISTS_CHANGED_EVENT, handler);
  return () => window.removeEventListener(PLAYLISTS_CHANGED_EVENT, handler);
}

export async function fetchPlaylists(options?: { force?: boolean }): Promise<UserPlaylist[]> {
  if (!options?.force && playlistsCache) {
    return playlistsCache;
  }
  const items = await ListPlaylists();
  playlistsCache = items
    .map((item) => normalizePlaylist(item))
    .filter((item) => item.id && item.name);
  for (const item of playlistsCache) {
    playlistDetailCache.set(item.id, item);
  }
  return playlistsCache;
}

export async function fetchPlaylist(
  id: string,
  options?: { force?: boolean },
): Promise<UserPlaylist | null> {
  const playlistId = id.trim();
  if (!playlistId) return null;

  if (!options?.force) {
    const cached = playlistDetailCache.get(playlistId);
    if (cached) return cached;
    if (playlistsCache) {
      const fromList = playlistsCache.find((item) => item.id === playlistId);
      if (fromList) return fromList;
    }
  }

  try {
    const item = await GetPlaylist(playlistId);
    const playlist = normalizePlaylist(item);
    if (!playlist.id) return null;
    playlistDetailCache.set(playlist.id, playlist);
    return playlist;
  } catch {
    return null;
  }
}

export async function createUserPlaylist(name: string): Promise<UserPlaylist> {
  const created = await CreatePlaylist(name.trim());
  notifyPlaylistsChanged();
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
  notifyPlaylistsChanged();
}

export async function removeSongFromPlaylist(playlistId: string, song: FavoriteSong): Promise<void> {
  await RemovePlaylistSong(playlistId, toWailsFavoriteSong(song));
  notifyPlaylistsChanged();
}

export async function deleteUserPlaylist(playlistId: string): Promise<void> {
  await DeletePlaylist(playlistId);
  notifyPlaylistsChanged();
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
