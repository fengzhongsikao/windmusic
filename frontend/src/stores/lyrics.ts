import { writable } from 'svelte/store';
import type { PlayerTrack } from '@/stores/player';
import { GetLyric } from '../../wailsjs/go/main/App';

export const lrcRaw = writable('');
export const lyricLoading = writable(false);
export const lyricError = writable('');

let lyricLoadToken = 0;

export function trackPlaybackKey(track: PlayerTrack): string {
  const ctx = track.playback;
  return `${track.id}|${ctx?.sourceId ?? ''}|${ctx?.platform ?? ''}|${ctx?.metaJson ?? ''}`;
}

export async function loadLyricsForTrack(track: PlayerTrack) {
  const token = ++lyricLoadToken;
  const ctx = track.playback;

  if (!ctx?.sourceId || !ctx.platform || !ctx.metaJson) {
    lyricLoading.set(false);
    lyricError.set('当前歌曲缺少歌词请求参数');
    lrcRaw.set('');
    return;
  }

  lyricLoading.set(true);
  lyricError.set('');
  try {
    const lyricInfo = await GetLyric(ctx.sourceId, ctx.platform, ctx.metaJson);
    if (token !== lyricLoadToken) {
      return;
    }
    const lyricText = (lyricInfo?.lyric ?? '').trim();
    if (!lyricText) {
      lyricError.set('未获取到歌词');
      lrcRaw.set('');
      return;
    }
    lrcRaw.set(lyricText);
    lyricError.set('');
  } catch (err) {
    if (token !== lyricLoadToken) {
      return;
    }
    lyricError.set(err instanceof Error ? err.message : String(err));
    lrcRaw.set('');
  } finally {
    if (token === lyricLoadToken) {
      lyricLoading.set(false);
    }
  }
}
