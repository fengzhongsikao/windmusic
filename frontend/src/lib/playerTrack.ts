import type { TrackItem } from '@/lib/track';
import type { PlayerTrack, PlaybackContext } from '@/stores/player';
import { music } from '../../wailsjs/go/models';

type SongItem = music.SongItem;

export function buildPlaybackContext(
  song: SongItem | undefined,
  sourceId: string,
  fallbackPlatform: string,
): PlaybackContext | undefined {
  if (!song || !sourceId) {
    return undefined;
  }
  return {
    sourceId,
    platform: song.source || fallbackPlatform,
    metaJson: song.metaJson,
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
    coverUrl: track.coverUrl,
    playback,
  };
}
