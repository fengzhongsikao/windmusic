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
} from '../../wailsjs/go/main/App';
import { music } from '../../wailsjs/go/models';
import type { PlayerTrack } from '@/stores/player.svelte';
import defaultCover from '@/assets/images/default.jpg';
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

export { Search, GetMusicURL, GetLyric, GetPic };
export { ListFavorites };

const FAVORITES_CHANGED_EVENT = 'favorites:changed';

function notifyFavoritesChanged() {
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent(FAVORITES_CHANGED_EVENT));
  }
}

export function onFavoritesChanged(listener: () => void): () => void {
  if (typeof window === 'undefined') {
    return () => {};
  }
  const handler = () => listener();
  window.addEventListener(FAVORITES_CHANGED_EVENT, handler);
  return () => window.removeEventListener(FAVORITES_CHANGED_EVENT, handler);
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
  if (!ctx?.sourceId || !ctx.metaJson) {
    return defaultCover;
  }

  try {
    const url = await GetPic(ctx.sourceId, ctx.platform, ctx.metaJson);
    return url?.trim() || defaultCover;
  } catch {
    return defaultCover;
  }
}

export function toFavoriteSong(track: PlayerTrack): FavoriteSong {
  return {
    id: String(track.id ?? ''),
    title: track.title ?? '',
    artist: track.artist ?? '',
    album: track.album ?? '',
    duration: track.duration ?? '',
    coverUrl: track.coverUrl ?? '',
    sourceId: track.playback?.sourceId ?? '',
    platform: track.playback?.platform ?? '',
    metaJson: track.playback?.metaJson ?? '',
  };
}

export async function checkTrackFavorite(track: PlayerTrack): Promise<boolean> {
  return IsFavorite(toFavoriteSong(track));
}

export async function addTrackFavorite(track: PlayerTrack): Promise<void> {
  await AddFavorite(toFavoriteSong(track));
  notifyFavoritesChanged();
}

export async function removeTrackFavorite(track: PlayerTrack): Promise<void> {
  await RemoveFavorite(toFavoriteSong(track));
  notifyFavoritesChanged();
}
