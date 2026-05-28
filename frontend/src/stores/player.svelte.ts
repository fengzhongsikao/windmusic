import type { TrackItem } from '@/lib/track';
import defaultCover from '@/assets/images/default.jpg';

export type ViewMode = 'normal' | 'immersive';

export type PlaybackContext = {
  sourceId: string;
  platform: string;
  metaJson: string;
};

export type PlayerTrack = Pick<TrackItem, 'id' | 'title' | 'artist' | 'album' | 'coverUrl'> & {
  playback?: PlaybackContext;
};

export const player = $state({
  viewMode: 'normal' as ViewMode,
  currentSong: {
    id: 'init-galaxy',
    title: '在银河中孤独摇摆',
    artist: '知更鸟 / HOYO-MiX',
    album: '未知专辑',
    coverUrl: defaultCover,
  } as PlayerTrack,
  isPlaying: false,
});

export function openImmersiveView() {
  player.viewMode = 'immersive';
}

export function closeImmersiveView() {
  player.viewMode = 'normal';
}

export function toggleImmersiveView() {
  player.viewMode = player.viewMode === 'immersive' ? 'normal' : 'immersive';
}

export function setCurrentTrack(track: PlayerTrack) {
  player.currentSong = track;
}

export function togglePlayByTrack(track: PlayerTrack) {
  if (String(player.currentSong.id) === String(track.id)) {
    player.isPlaying = !player.isPlaying;
    return;
  }
  player.currentSong = track;
  player.isPlaying = true;
}

export function togglePlayerPlayback() {
  player.isPlaying = !player.isPlaying;
}

export function setPlaying(playing: boolean) {
  player.isPlaying = playing;
}
