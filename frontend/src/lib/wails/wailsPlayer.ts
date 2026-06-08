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
  RecordRecent,
  RemoveRecent,
  ClearRecent,
} from '../../../wailsjs/go/main/App';
import { music } from '../../../wailsjs/go/models';
import type { PlayerTrack } from '@/stores/playback/player.svelte';
import defaultCover from '@/assets/images/default.jpg';
import { fetchLocalSongExtras, isLocalStoredSong, LOCAL_PLATFORM, localPathFromMetaJson } from '@/lib/library/localMusic';
import { getMetingURL, metingSourceId } from '@/stores/sources/meting.svelte';
import {
  favoriteSongKey,
  normalizeFavoriteSong,
  sameFavoriteSong,
  type FavoriteSong,
} from '@/lib/library/favoriteSong';
import {
  getFavorites,
  refreshFavoritesFromBackend,
} from '@/stores/library/favorites.svelte';
import { FAVORITES_UPDATED_EVENT } from '@/stores/library/favorites.svelte';
import { RECENT_UPDATED_EVENT } from '@/stores/library/recent.svelte';
import { trackPlaybackKey } from '@/stores/playback/lyrics';
import { EventsOn } from '../../../wailsjs/runtime/runtime';

export type SongItem = music.SongItem;
export type SearchResult = music.SearchResult;
export type { FavoriteSong } from '@/lib/library/favoriteSong';
export type { RecentSong } from '@/lib/library/recentSong';
export { favoriteSongKey, normalizeFavoriteSong } from '@/lib/library/favoriteSong';
export { normalizeRecentSong } from '@/lib/library/recentSong';

export { Search, GetMusicURL, GetLyric, GetPic };

/** 读取收藏；默认用内存状态，force 时从 Go 重新拉取 */
export async function fetchFavorites(options?: { force?: boolean }): Promise<FavoriteSong[]> {
  if (!options?.force && getFavorites().length > 0) {
    return [...getFavorites()];
  }
  const items = await ListFavorites();
  return items.map((item) => normalizeFavoriteSong(item));
}

export function onFavoritesChanged(listener: () => void): () => void {
  return EventsOn(FAVORITES_UPDATED_EVENT, () => listener());
}

export function onRecentChanged(listener: () => void): () => void {
  return EventsOn(RECENT_UPDATED_EVENT, () => listener());
}

import { getRecentSongs } from '@/stores/library/recent.svelte';

export async function fetchRecentSongs(): Promise<import('@/lib/library/recentSong').RecentSong[]> {
  return [...getRecentSongs()];
}

export async function resolveReadySourceId(): Promise<string> {
  const metingURL = getMetingURL();
  if (metingURL) {
    return metingSourceId(metingURL);
  }
  throw new Error('请先在设置中配置 Meting 源');
}

/** 通过音源获取封面 URL（网络或本地路径，由 Go 端返回） */
const coverUrlInflight = new Map<string, Promise<string>>();

export async function fetchCoverUrl(track: PlayerTrack): Promise<string> {
  const existing = track.coverUrl?.trim();
  if (existing) {
    return existing;
  }

  const ctx = track.playback;
  if (!ctx) {
    return defaultCover;
  }

  const cacheKey = trackPlaybackKey(track);
  const inflight = coverUrlInflight.get(cacheKey);
  if (inflight) {
    return inflight;
  }

  const promise = fetchCoverUrlUncached(track).finally(() => {
    coverUrlInflight.delete(cacheKey);
  });
  coverUrlInflight.set(cacheKey, promise);
  return promise;
}

async function fetchCoverUrlUncached(track: PlayerTrack): Promise<string> {
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

const EMPTY_FIELD_MARKERS = new Set(['—', '-', '未知专辑', 'unknown']);

function emptyIfUnknown(value: string | undefined): string {
  const trimmed = value?.trim() ?? '';
  if (!trimmed || EMPTY_FIELD_MARKERS.has(trimmed)) {
    return '';
  }
  return trimmed;
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

function shouldOmitStoredCoverUrl(track: PlayerTrack): boolean {
  const ctx = track.playback;
  if (isLocalStoredSong({ platform: ctx?.platform, sourceId: ctx?.sourceId })) {
    return true;
  }
  const url = track.coverUrl?.trim() ?? '';
  return (
    url.startsWith('data:') ||
    /^https?:\/\/127\.0\.0\.1(?::\d+)?/i.test(url) ||
    /^https?:\/\/localhost(?::\d+)?/i.test(url)
  );
}

export function toFavoriteSong(track: PlayerTrack): FavoriteSong {
  const { album, duration } = albumAndDurationFromTrack(track);
  return {
    id: String(track.id ?? ''),
    title: track.title ?? '',
    artist: track.artist ?? '',
    album,
    duration,
    coverUrl: shouldOmitStoredCoverUrl(track) ? '' : (track.coverUrl ?? ''),
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
  const cached = getFavorites();
  if (cached.length > 0) {
    return cached.some((item) => sameFavoriteSong(item, song));
  }
  return IsFavorite(toWailsFavoriteSong(song));
}

export async function addTrackFavorite(track: PlayerTrack): Promise<void> {
  await AddFavorite(toWailsFavoriteSong(toFavoriteSong(track)));
  await refreshFavoritesFromBackend();
}

export async function removeTrackFavorite(track: PlayerTrack): Promise<void> {
  await RemoveFavorite(toWailsFavoriteSong(toFavoriteSong(track)));
  await refreshFavoritesFromBackend();
}

export function toRecentSong(track: PlayerTrack): import('@/lib/library/recentSong').RecentSong {
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
  } catch (err) {
    console.warn('[最近播放] 记录失败', err);
  }
}

export async function removeRecentSong(song: import('@/lib/library/recentSong').RecentSong): Promise<void> {
  await RemoveRecent(toWailsRecentSong({ ...song }));
}

export async function clearRecentHistory(): Promise<void> {
  await ClearRecent();
}
