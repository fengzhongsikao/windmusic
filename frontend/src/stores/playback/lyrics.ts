import { writable } from 'svelte/store';
import type { PlayerTrack } from '@/stores/playback/player';
import { GetLyric } from '../../../wailsjs/go/main/App';
import { fetchLocalSongExtras } from '@/lib/localMusic';

export const lrcRaw = writable('');
export const lyricLoading = writable(false);
export const lyricError = writable('');

let lyricLoadToken = 0;

export function trackPlaybackKey(track: PlayerTrack): string {
  const ctx = track.playback;
  return `${track.id}|${ctx?.localPath ?? ''}|${ctx?.sourceId ?? ''}|${ctx?.platform ?? ''}|${ctx?.metaJson ?? ''}`;
}

function lyricFromLocalMeta(metaJson: string): string {
  if (!metaJson.trim()) {
    return '';
  }
  try {
    const meta = JSON.parse(metaJson) as { lyric?: string };
    return (meta.lyric ?? '').trim();
  } catch {
    return '';
  }
}

export async function loadLyricsForTrack(track: PlayerTrack) {
  const token = ++lyricLoadToken;
  const ctx = track.playback;

  if (ctx?.localPath) {
    lyricLoading.set(true);
    lyricError.set('');
    let lyricText = lyricFromLocalMeta(ctx.metaJson ?? '');
    if (!lyricText) {
      const extras = await fetchLocalSongExtras(ctx.localPath);
      if (token !== lyricLoadToken) {
        return;
      }
      lyricText = extras.lyric;
    }
    if (token !== lyricLoadToken) {
      return;
    }
    if (lyricText) {
      lrcRaw.set(lyricText);
      lyricError.set('');
    } else {
      lrcRaw.set('');
      lyricError.set('未找到本地歌词');
    }
    lyricLoading.set(false);
    return;
  }

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
