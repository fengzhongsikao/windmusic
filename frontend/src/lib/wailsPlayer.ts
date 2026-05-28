/**
 * Wails 播放器相关 API 封装（绑定自 wailsjs/go/main/App）。
 */
import {
  Search,
  ListSources,
  GetMusicURL,
  GetLyric,
  GetPic,
} from '../../wailsjs/go/main/App';
import { music } from '../../wailsjs/go/models';
import type { PlayerTrack } from '@/stores/player.svelte';
import defaultCover from '@/assets/images/default.jpg';

export type SongItem = music.SongItem;
export type SearchResult = music.SearchResult;

export { Search, ListSources, GetMusicURL, GetLyric, GetPic };

export async function resolveReadySourceId(): Promise<string> {
  const sources = await ListSources();
  const ready = sources.find((item) => item.enabled && item.status === 'ready');
  if (!ready) {
    throw new Error('请先在设置中导入并启用音源');
  }
  return ready.id;
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
