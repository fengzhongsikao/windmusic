import { writable } from 'svelte/store';
import { GetLyric } from '../../wailsjs/go/main/App';
import type { PlayerTrack } from '@/stores/player';

const DEMO_LRC = `[00:00.00]在银河中孤独摇摆
[00:08.50]当星光落在窗台
[00:16.20]我听见宇宙在轻声对白
[00:24.00]把沉默都唱成期待
[00:31.80]在银河中孤独摇摆
[00:39.50]让旋律穿过云海
[00:47.20]每一句歌词都是告白
[00:55.00]在夜里把梦点亮`;

export const lrcRaw = writable(DEMO_LRC);
export const lyricLoading = writable(false);
export const lyricError = writable('');

let lyricLoadToken = 0;

function errorMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}

export function trackPlaybackKey(track: PlayerTrack): string {
  const ctx = track.playback;
  return `${track.id}|${ctx?.sourceId ?? ''}|${ctx?.platform ?? ''}|${ctx?.metaJson ?? ''}`;
}

export async function loadLyricsForTrack(track: PlayerTrack) {
  const ctx = track.playback;
  const token = ++lyricLoadToken;

  if (!ctx?.sourceId || !ctx.platform || !ctx.metaJson) {
    lrcRaw.set(DEMO_LRC);
    lyricError.set('');
    lyricLoading.set(false);
    return;
  }

  lyricLoading.set(true);
  lyricError.set('');

  try {
    const info = await GetLyric(ctx.sourceId, ctx.platform, ctx.metaJson);
    if (token !== lyricLoadToken) {
      return;
    }

    const text = info.lyric?.trim() || info.tlyric?.trim() || '';
    if (text) {
      lrcRaw.set(text);
    } else {
      lrcRaw.set(DEMO_LRC);
      lyricError.set('未获取到歌词，已显示示例歌词');
    }
  } catch (err) {
    if (token !== lyricLoadToken) {
      return;
    }
    lrcRaw.set(DEMO_LRC);
    lyricError.set(errorMessage(err));
  } finally {
    if (token === lyricLoadToken) {
      lyricLoading.set(false);
    }
  }
}
