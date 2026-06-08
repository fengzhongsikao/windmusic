import type { TrackItem } from '@/lib/playback/track';

export type LocalSortOption =
  | 'title-asc'
  | 'duration-asc'
  | 'duration-desc'
  | 'size-asc'
  | 'size-desc';

export const LOCAL_SORT_OPTIONS: { value: LocalSortOption; label: string }[] = [
  { value: 'title-asc', label: '标题 (A-Z)' },
  { value: 'duration-asc', label: '时长（短→长）' },
  { value: 'duration-desc', label: '时长（长→短）' },
  { value: 'size-asc', label: '大小（小→大）' },
  { value: 'size-desc', label: '大小（大→小）' },
];

const EMPTY_MARKERS = new Set(['', '—', '-']);

export function parseDurationSeconds(label: string | undefined): number {
  const trimmed = label?.trim() ?? '';
  if (EMPTY_MARKERS.has(trimmed)) {
    return -1;
  }
  const parts = trimmed.split(':').map((part) => Number.parseInt(part, 10));
  if (parts.some((n) => Number.isNaN(n))) {
    return -1;
  }
  if (parts.length === 2) {
    return parts[0] * 60 + parts[1];
  }
  if (parts.length === 3) {
    return parts[0] * 3600 + parts[1] * 60 + parts[2];
  }
  return -1;
}

const SIZE_UNITS: Record<string, number> = {
  B: 1,
  KB: 1024,
  MB: 1024 ** 2,
  GB: 1024 ** 3,
  TB: 1024 ** 4,
};

export function parseFileSizeBytes(label: string | undefined): number {
  const trimmed = label?.trim() ?? '';
  if (EMPTY_MARKERS.has(trimmed)) {
    return -1;
  }
  const match = trimmed.match(/^([\d.]+)\s*(B|KB|MB|GB|TB)$/i);
  if (!match) {
    return -1;
  }
  const value = Number.parseFloat(match[1]);
  const unit = match[2].toUpperCase();
  if (Number.isNaN(value)) {
    return -1;
  }
  return value * (SIZE_UNITS[unit] ?? 1);
}

function compareTitle(left: TrackItem, right: TrackItem): number {
  const byTitle = left.title.localeCompare(right.title, 'zh-CN', { sensitivity: 'base' });
  if (byTitle !== 0) {
    return byTitle;
  }
  return String(left.listKey ?? left.id).localeCompare(String(right.listKey ?? right.id));
}

export function sortLocalTracks(tracks: TrackItem[], option: LocalSortOption): TrackItem[] {
  if (option === 'title-asc') {
    return tracks;
  }

  const sorted = [...tracks];
  const dir = option.endsWith('-desc') ? -1 : 1;

  if (option.startsWith('duration')) {
    sorted.sort((a, b) => {
      const left = parseDurationSeconds(a.duration);
      const right = parseDurationSeconds(b.duration);
      if (left !== right) {
        return (left - right) * dir;
      }
      return compareTitle(a, b);
    });
    return sorted;
  }

  if (option.startsWith('size')) {
    sorted.sort((a, b) => {
      const left = parseFileSizeBytes(a.size);
      const right = parseFileSizeBytes(b.size);
      if (left !== right) {
        return (left - right) * dir;
      }
      return compareTitle(a, b);
    });
    return sorted;
  }

  return sorted;
}
