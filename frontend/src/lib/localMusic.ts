import {
  GetLocalAudioStream,
  GetLocalFolderSongs,
  GetLocalLibrarySnapshot,
  GetLocalLibraryTracksIndex,
  GetLocalSongCovers,
  GetLocalSongExtras,
  PickLocalMusicFolder,
  RemoveLocalMusicFolder,
  ScanLocalLibrary,
  SetLocalFolderAlias,
} from '../../wailsjs/go/main/App';
import { music } from '../../wailsjs/go/models';
import type { TrackItem } from '@/lib/track';
import type { PlaybackContext, PlayerTrack } from '@/stores/player.svelte';
import { trackItemToPlayerTrack } from '@/lib/playerTrack';

export type LocalSong = music.LocalSong;

export const LOCAL_PLATFORM = 'local';
export const LOCAL_SOURCE_ID = 'local';

export {
  PickLocalMusicFolder,
  GetLocalFolderSongs,
  GetLocalLibrarySnapshot,
  GetLocalLibraryTracksIndex,
  RemoveLocalMusicFolder,
  ScanLocalLibrary,
  SetLocalFolderAlias,
  GetLocalAudioStream,
  GetLocalSongExtras,
};

const COVER_BATCH_SIZE = 80;

export type LocalCoverBatch = {
  covers: Record<string, string>;
  paths: Record<string, string>;
};

export function applyLocalCoverBatch(
  batch: LocalCoverBatch,
  target: Record<string, string>,
): void {
  for (const [path, key] of Object.entries(batch.paths)) {
    const cover = batch.covers[key];
    if (cover) {
      target[path] = cover;
    }
  }
}

function mergeCoverBatches(left: LocalCoverBatch, right: LocalCoverBatch): LocalCoverBatch {
  return {
    covers: { ...left.covers, ...right.covers },
    paths: { ...left.paths, ...right.paths },
  };
}

function normalizeCoverBatch(raw: {
  covers?: Record<string, string>;
  paths?: Record<string, string>;
}): LocalCoverBatch {
  return {
    covers: raw.covers ?? {},
    paths: raw.paths ?? {},
  };
}

export async function fetchLocalSongCovers(filePaths: string[]): Promise<LocalCoverBatch> {
  const paths = [...new Set(filePaths.map((p) => p.trim()).filter(Boolean))];
  if (paths.length === 0) {
    return { covers: {}, paths: {} };
  }

  let merged: LocalCoverBatch = { covers: {}, paths: {} };
  for (let i = 0; i < paths.length; i += COVER_BATCH_SIZE) {
    const chunk = paths.slice(i, i + COVER_BATCH_SIZE);
    try {
      const batch = normalizeCoverBatch(await GetLocalSongCovers(chunk));
      merged = mergeCoverBatches(merged, batch);
    } catch {
      // skip failed batch
    }
  }
  return merged;
}

export async function fetchLocalSongExtras(filePath: string) {
  const path = filePath.trim();
  if (!path) {
    return { coverData: '', lyric: '' };
  }
  try {
    const extras = await GetLocalSongExtras(path);
    return {
      coverData: extras?.coverData?.trim() ?? '',
      lyric: extras?.lyric?.trim() ?? '',
    };
  } catch {
    return { coverData: '', lyric: '' };
  }
}

export type StoredPlaybackSong = {
  id?: string;
  sourceId?: string;
  platform?: string;
  metaJson?: string;
};

export function isLocalStoredSong(song: StoredPlaybackSong): boolean {
  const platform = song.platform?.trim() ?? '';
  const sourceId = song.sourceId?.trim() ?? '';
  return platform === LOCAL_PLATFORM || sourceId === LOCAL_SOURCE_ID;
}

export function localPathFromMetaJson(metaJson?: string): string {
  if (!metaJson?.trim()) {
    return '';
  }
  try {
    const meta = JSON.parse(metaJson) as { filePath?: string };
    return meta.filePath?.trim() ?? '';
  } catch {
    return '';
  }
}

/** 从收藏/最近播放/歌单等持久化记录还原播放上下文 */
export function buildPlaybackContextFromStored(song: StoredPlaybackSong): PlaybackContext | undefined {
  const sourceId = song.sourceId?.trim() ?? '';
  const platform = song.platform?.trim() ?? '';
  const metaJson = song.metaJson ?? '';

  if (!sourceId && !platform && !metaJson) {
    return undefined;
  }

  if (isLocalStoredSong(song)) {
    const localPath = localPathFromMetaJson(metaJson) || song.id?.trim() || '';
    if (!localPath) {
      return undefined;
    }
    return {
      sourceId: LOCAL_SOURCE_ID,
      platform: LOCAL_PLATFORM,
      localPath,
      metaJson,
    };
  }

  return { sourceId, platform, metaJson };
}

export function storedSongToPlayerTrack(song: StoredPlaybackSong & {
  id: string;
  title: string;
  artist: string;
  album?: string;
  duration?: string;
  coverUrl?: string;
}): PlayerTrack {
  return {
    id: song.id,
    title: song.title,
    artist: song.artist,
    album: song.album ?? '',
    duration: song.duration?.trim() || '—',
    coverUrl: song.coverUrl?.trim() || undefined,
    playback: buildPlaybackContextFromStored(song),
  };
}

export function buildLocalPlaybackContext(
  song: LocalSong,
  extras?: { lyric?: string },
): PlaybackContext {
  return {
    sourceId: LOCAL_SOURCE_ID,
    platform: LOCAL_PLATFORM,
    localPath: song.filePath,
    metaJson: JSON.stringify({
      album: song.album ?? '',
      duration: song.duration ?? '',
      lyric: extras?.lyric ?? song.lyric ?? '',
      filePath: song.filePath,
    }),
  };
}

export function localSongToTrackItem(song: LocalSong): TrackItem {
  const cover = song.coverData?.trim();
  return {
    id: song.id,
    listKey: song.filePath,
    title: song.title,
    artist: song.artist,
    album: song.album ?? '—',
    duration: song.duration?.trim() || '—',
    size: song.size?.trim() || '—',
    coverUrl: cover || undefined,
  };
}

export function localSongToPlayerTrack(song: LocalSong, coverUrl?: string): PlayerTrack {
  const track = trackItemToPlayerTrack(localSongToTrackItem(song), buildLocalPlaybackContext(song));
  if (coverUrl?.trim()) {
    track.coverUrl = coverUrl.trim();
  }
  return track;
}

export async function localSongToPlayerTrackAsync(song: LocalSong): Promise<PlayerTrack> {
  const { coverData, lyric } = await fetchLocalSongExtras(song.filePath);
  const item = localSongToTrackItem(song);
  if (coverData) {
    item.coverUrl = coverData;
  }
  return trackItemToPlayerTrack(item, buildLocalPlaybackContext(song, { lyric }));
}

export function songInFolder(filePath: string, folderPath: string): boolean {
  const songPath = filePath.replace(/\\/g, '/');
  const folder = folderPath.replace(/\\/g, '/').replace(/\/$/, '');
  return songPath === folder || songPath.startsWith(`${folder}/`);
}

export function folderDisplayName(path: string): string {
  const trimmed = path.trim();
  if (!trimmed) {
    return path;
  }
  const parts = trimmed.split(/[/\\]/).filter(Boolean);
  return parts[parts.length - 1] ?? trimmed;
}

export function folderDisplayLabel(path: string, aliases?: Record<string, string>): string {
  const alias = aliases?.[path]?.trim();
  if (alias) {
    return alias;
  }
  return folderDisplayName(path);
}
