/**
 * Wails 播放器相关 API 封装（绑定自 wailsjs/go/main/App）。
 */
import {
  Search,
  GetMusicURL,
  GetLyric,
  GetPic,
  ListFavorites,
  AddFavorite,
  RemoveFavorite,
  IsFavorite,
  ListRecent,
  RecordRecent,
  RemoveRecent,
  ClearRecent,
} from '../../wailsjs/go/main/App';
import { music } from '../../wailsjs/go/models';
import type { PlayerTrack } from '@/stores/player.svelte';
import defaultCover from '@/assets/images/default.jpg';
import { fetchLocalSongExtras, isLocalStoredSong, LOCAL_PLATFORM, localPathFromMetaJson } from '@/lib/localMusic';
import { getMetingURL, metingSourceId } from '@/lib/meting';

export type SongItem = music.SongItem;
export type SearchResult = music.SearchResult;
export type FavoriteSong = {
  id: string;
  title: string;
  artist: string;
  album?: string;
  duration?: string;
  coverUrl?: string;
  sourceId?: string;
  platform?: string;
  metaJson?: string;
};

export type RecentSong = FavoriteSong & {
  playedAt: string;
};

export { Search, GetMusicURL, GetLyric, GetPic };
export { ListFavorites, ListRecent };

const FAVORITES_CHANGED_EVENT = 'favorites:changed';
const RECENT_CHANGED_EVENT = 'recent:changed';

let favoritesListCache: FavoriteSong[] | null = null;

const EMPTY_FIELD_MARKERS = new Set(['—', '-', '未知专辑', 'unknown']);

function emptyIfUnknown(value: string | undefined): string {
  const trimmed = value?.trim() ?? '';
  if (!trimmed || EMPTY_FIELD_MARKERS.has(trimmed)) {
    return '';
  }
  return trimmed;
}

function normalizeFavoriteSong(raw: {
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
  let album = emptyIfUnknown(raw.album);
  let duration = emptyIfUnknown(raw.duration);
  const metaJson = raw.metaJson ?? '';

  if ((!album || !duration) && metaJson) {
    try {
      const meta = JSON.parse(metaJson) as {
        album?: string;
        duration?: string;
        interval?: string;
      };
      if (!album) {
        album = emptyIfUnknown(meta.album);
      }
      if (!duration) {
        duration = emptyIfUnknown(meta.duration ?? meta.interval);
      }
    } catch {
      // ignore malformed metaJson
    }
  }

  return {
    id: String(raw.id ?? ''),
    title: raw.title ?? '',
    artist: raw.artist ?? '',
    album,
    duration,
    coverUrl: raw.coverUrl ?? '',
    sourceId: raw.sourceId ?? '',
    platform: raw.platform ?? '',
    metaJson,
  };
}

export function favoriteSongKey(song: FavoriteSong): string {
  const entry = normalizeFavoriteSong(song);
  return [entry.id, entry.sourceId, entry.platform, entry.metaJson].join('|');
}

function sameFavoriteSong(a: FavoriteSong, b: FavoriteSong): boolean {
  if (a.metaJson && b.metaJson) {
    if (a.metaJson !== b.metaJson) return false;
    if (a.platform && b.platform && a.platform !== b.platform) return false;
    if (a.sourceId && b.sourceId && a.sourceId !== b.sourceId) return false;
    return true;
  }
  if (a.id && b.id) {
    if (String(a.id) !== String(b.id)) return false;
    if (a.platform && b.platform && a.platform !== b.platform) return false;
    if (a.sourceId && b.sourceId && a.sourceId !== b.sourceId) return false;
    return true;
  }
  return Boolean(a.title && b.title && a.artist && b.artist && a.title === b.title && a.artist === b.artist);
}

function notifyFavoritesChanged() {
  favoritesListCache = null;
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent(FAVORITES_CHANGED_EVENT));
  }
}

/** 读取收藏列表（带内存缓存，避免同页重复走 Wails 读盘） */
export async function fetchFavorites(options?: { force?: boolean }): Promise<FavoriteSong[]> {
  if (!options?.force && favoritesListCache) {
    return favoritesListCache;
  }
  const items = await ListFavorites();
  const seen = new Set<string>();
  favoritesListCache = items
    .map((item) => normalizeFavoriteSong(item))
    .filter((item) => {
      const key = favoriteSongKey(item);
      if (seen.has(key)) return false;
      seen.add(key);
      return true;
    });
  return favoritesListCache;
}

export function onFavoritesChanged(listener: () => void): () => void {
  if (typeof window === 'undefined') {
    return () => {};
  }
  const handler = () => listener();
  window.addEventListener(FAVORITES_CHANGED_EVENT, handler);
  return () => window.removeEventListener(FAVORITES_CHANGED_EVENT, handler);
}

function notifyRecentChanged() {
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent(RECENT_CHANGED_EVENT));
  }
}

export function onRecentChanged(listener: () => void): () => void {
  if (typeof window === 'undefined') {
    return () => {};
  }
  const handler = () => listener();
  window.addEventListener(RECENT_CHANGED_EVENT, handler);
  return () => window.removeEventListener(RECENT_CHANGED_EVENT, handler);
}

function playedAtToIso(value: unknown): string {
  if (typeof value === 'string') {
    return value;
  }
  if (value instanceof Date) {
    return value.toISOString();
  }
  return '';
}

export function normalizeRecentSong(raw: {
  id?: string;
  title?: string;
  artist?: string;
  album?: string;
  duration?: string;
  coverUrl?: string;
  sourceId?: string;
  platform?: string;
  metaJson?: string;
  playedAt?: unknown;
}): RecentSong {
  return {
    id: String(raw.id ?? ''),
    title: raw.title ?? '',
    artist: raw.artist ?? '',
    album: raw.album ?? '',
    duration: raw.duration ?? '',
    coverUrl: raw.coverUrl ?? '',
    sourceId: raw.sourceId ?? '',
    platform: raw.platform ?? '',
    metaJson: raw.metaJson ?? '',
    playedAt: playedAtToIso(raw.playedAt),
  };
}

export async function fetchRecentSongs(): Promise<RecentSong[]> {
  const items = await ListRecent();
  return items.map((item) => normalizeRecentSong(item));
}

export async function resolveReadySourceId(): Promise<string> {
  const metingURL = getMetingURL();
  if (metingURL) {
    return metingSourceId(metingURL);
  }
  throw new Error('请先在设置中配置 Meting 源');
}

/** 通过音源获取封面 URL（网络或本地路径，由 Go 端返回） */
export async function fetchCoverUrl(track: PlayerTrack): Promise<string> {
  const existing = track.coverUrl?.trim();
  if (existing) {
    return existing;
  }

  const ctx = track.playback;
  if (!ctx) {
    return defaultCover;
  }

  if (isLocalStoredSong(ctx) || ctx.platform === LOCAL_PLATFORM) {
    const filePath = ctx.localPath?.trim() || localPathFromMetaJson(ctx.metaJson) || String(track.id ?? '');
    if (filePath) {
      const { coverData } = await fetchLocalSongExtras(filePath);
      return coverData || defaultCover;
    }
    return defaultCover;
  }

  if (!ctx.sourceId || !ctx.metaJson) {
    return defaultCover;
  }

  try {
    const url = await GetPic(ctx.sourceId, ctx.platform, ctx.metaJson);
    return url?.trim() || defaultCover;
  } catch {
    return defaultCover;
  }
}

function albumAndDurationFromTrack(track: PlayerTrack): { album: string; duration: string } {
  let album = emptyIfUnknown(track.album);
  let duration = emptyIfUnknown(track.duration);

  if ((!album || !duration) && track.playback?.metaJson) {
    try {
      const meta = JSON.parse(track.playback.metaJson) as {
        album?: string;
        duration?: string;
        interval?: string;
      };
      if (!album) {
        album = emptyIfUnknown(meta.album);
      }
      if (!duration) {
        duration = emptyIfUnknown(meta.duration ?? meta.interval);
      }
    } catch {
      // ignore malformed metaJson
    }
  }

  return { album, duration };
}

export function toFavoriteSong(track: PlayerTrack): FavoriteSong {
  const { album, duration } = albumAndDurationFromTrack(track);
  return {
    id: String(track.id ?? ''),
    title: track.title ?? '',
    artist: track.artist ?? '',
    album,
    duration,
    coverUrl: track.coverUrl ?? '',
    sourceId: track.playback?.sourceId ?? '',
    platform: track.playback?.platform ?? '',
    metaJson: track.playback?.metaJson ?? '',
  };
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

function toWailsRecentSong(song: FavoriteSong): music.RecentSong {
  return music.RecentSong.createFrom({
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

export async function checkTrackFavorite(track: PlayerTrack): Promise<boolean> {
  const song = toFavoriteSong(track);
  if (favoritesListCache) {
    return favoritesListCache.some((item) => sameFavoriteSong(item, song));
  }
  return IsFavorite(toWailsFavoriteSong(song));
}

export async function addTrackFavorite(track: PlayerTrack): Promise<void> {
  await AddFavorite(toWailsFavoriteSong(toFavoriteSong(track)));
  notifyFavoritesChanged();
}

export async function removeTrackFavorite(track: PlayerTrack): Promise<void> {
  await RemoveFavorite(toWailsFavoriteSong(toFavoriteSong(track)));
  notifyFavoritesChanged();
}

export function toRecentSong(track: PlayerTrack): RecentSong {
  return {
    ...toFavoriteSong(track),
    playedAt: '',
  };
}

function hasPlayback(track: PlayerTrack): boolean {
  const ctx = track.playback;
  if (ctx?.localPath?.trim()) {
    return true;
  }
  return Boolean(ctx?.sourceId && ctx.platform && ctx.metaJson);
}

export async function recordRecentPlay(track: PlayerTrack): Promise<void> {
  if (!hasPlayback(track)) {
    return;
  }
  try {
    await RecordRecent(toWailsRecentSong(toFavoriteSong(track)));
    notifyRecentChanged();
  } catch (err) {
    console.warn('[最近播放] 记录失败', err);
  }
}

export async function removeRecentSong(song: RecentSong): Promise<void> {
  await RemoveRecent(toWailsRecentSong({ ...song }));
  notifyRecentChanged();
}

export async function clearRecentHistory(): Promise<void> {
  await ClearRecent();
  notifyRecentChanged();
}
