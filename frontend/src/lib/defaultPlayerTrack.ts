import { getMetingURL, metingSourceId } from '@/stores/meting.svelte';
import type { PlayerTrack } from '@/stores/player';

export const DEFAULT_TRACK_ID = '2155422574';

const DEFAULT_METING_BASE = 'https://meting.mikus.ink';

/** 首页底部默认展示/试播曲目（Meting 网易云） */
export const DEFAULT_TRACK_META = {
  title: '在银河中孤独摇摆（伴奏）',
  author: 'HOYO-MiX',
  pic: 'http://p1.music.126.net/aR4BlDNkA84tFbg8bBpriA==/109951169585655912.jpg',
  url: `${DEFAULT_METING_BASE}/api?server=netease&type=url&id=${DEFAULT_TRACK_ID}`,
  lrc: `${DEFAULT_METING_BASE}/api?server=netease&type=lrc&id=${DEFAULT_TRACK_ID}`,
} as const;

function buildDefaultMetaJson(): string {
  return JSON.stringify({
    title: DEFAULT_TRACK_META.title,
    author: DEFAULT_TRACK_META.author,
    pic: DEFAULT_TRACK_META.pic,
    url: DEFAULT_TRACK_META.url,
    lrc: DEFAULT_TRACK_META.lrc,
    id: DEFAULT_TRACK_ID,
    server: 'netease',
  });
}

export function resolveDefaultMetingSourceId(): string {
  const base = getMetingURL() || DEFAULT_METING_BASE;
  return metingSourceId(base);
}

export function createDefaultPlayerTrack(): PlayerTrack {
  return {
    id: DEFAULT_TRACK_ID,
    title: DEFAULT_TRACK_META.title,
    artist: DEFAULT_TRACK_META.author,
    album: '',
    duration: '',
    coverUrl: DEFAULT_TRACK_META.pic,
    playback: {
      sourceId: resolveDefaultMetingSourceId(),
      platform: 'netease',
      metaJson: buildDefaultMetaJson(),
    },
  };
}
