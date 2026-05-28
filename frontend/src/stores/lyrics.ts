import { writable } from 'svelte/store';
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

export function trackPlaybackKey(track: PlayerTrack): string {
  const ctx = track.playback;
  return `${track.id}|${ctx?.sourceId ?? ''}|${ctx?.platform ?? ''}|${ctx?.metaJson ?? ''}`;
}

export async function loadLyricsForTrack(track: PlayerTrack) {
  const token = ++lyricLoadToken;

  lyricLoading.set(false);
  if (token !== lyricLoadToken) {
    return;
  }
  lrcRaw.set(DEMO_LRC);
  lyricError.set('');
}
