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

const EMPTY_FIELD_MARKERS = new Set(['—', '-', '未知专辑', 'unknown']);

function emptyIfUnknown(value: string | undefined): string {
  const trimmed = value?.trim() ?? '';
  if (!trimmed || EMPTY_FIELD_MARKERS.has(trimmed)) {
    return '';
  }
  return trimmed;
}

export function normalizeFavoriteSong(raw: {
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

export function sameFavoriteSong(a: FavoriteSong, b: FavoriteSong): boolean {
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
