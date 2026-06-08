import { normalizeFavoriteSong, type FavoriteSong } from '@/lib/library/favoriteSong';

export type RecentSong = FavoriteSong & {
  playedAt: string;
};

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
    ...normalizeFavoriteSong(raw),
    playedAt: playedAtToIso(raw.playedAt),
  };
}
