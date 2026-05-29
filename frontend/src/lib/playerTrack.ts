import type { TrackItem } from '@/lib/track';
import type { PlayerTrack, PlaybackContext } from '@/stores/player';
import { music } from '../../wailsjs/go/models';

type SongItem = music.SongItem;

/** 将专辑、时长写入 metaJson，收藏/最近播放可回读 */
export function withTrackMetaFields(
  metaJson: string,
  fields: { album?: string; duration?: string },
): string {
  const album = fields.album?.trim() ?? '';
  const duration = fields.duration?.trim() ?? '';
  if (!album && !duration) {
    return metaJson;
  }
  try {
    const parsed = JSON.parse(metaJson) as Record<string, unknown>;
    if (album) {
      parsed.album = album;
    }
    if (duration) {
      parsed.duration = duration;
    }
    return JSON.stringify(parsed);
  } catch {
    return metaJson;
  }
}

export function buildPlaybackContext(
  song: SongItem | undefined,
  sourceId: string,
  fallbackPlatform: string,
): PlaybackContext | undefined {
  if (!song || !sourceId) {
    return undefined;
  }
  const metaJson = withTrackMetaFields(song.metaJson, {
    album: song.album,
    duration: song.interval,
  });
  return {
    sourceId,
    platform: song.source || fallbackPlatform,
    metaJson,
  };
}

export function trackItemToPlayerTrack(
  track: TrackItem,
  playback?: PlaybackContext,
): PlayerTrack {
  return {
    id: track.id,
    title: track.title,
    artist: track.artist,
    album: track.album,
    duration: track.duration,
    coverUrl: track.coverUrl,
    playback,
  };
}
