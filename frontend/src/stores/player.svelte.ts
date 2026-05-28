import type { TrackItem } from '@/lib/track';
import defaultCover from '@/assets/images/default.jpg';

export type ViewMode = 'normal' | 'immersive';

export type PlaybackContext = {
  sourceId: string;
  platform: string;
  metaJson: string;
};

export type PlayerTrack = Pick<
  TrackItem,
  'id' | 'title' | 'artist' | 'album' | 'duration' | 'coverUrl'
> & {
  playback?: PlaybackContext;
};

export type RepeatMode = 'off' | 'all' | 'one';

const DEFAULT_VOLUME = 30;
const SETTINGS_KEY = 'windmusic:player-settings';

function sameTrack(a: PlayerTrack, b: PlayerTrack): boolean {
  return (
    String(a.id) === String(b.id) &&
    (a.playback?.sourceId ?? '') === (b.playback?.sourceId ?? '') &&
    (a.playback?.platform ?? '') === (b.playback?.platform ?? '') &&
    (a.playback?.metaJson ?? '') === (b.playback?.metaJson ?? '')
  );
}

function readPersistedSettings(): {
  volume: number;
  muted: boolean;
  repeatMode: RepeatMode;
  shuffled: boolean;
} {
  if (typeof window === 'undefined') {
    return { volume: DEFAULT_VOLUME, muted: false, repeatMode: 'off', shuffled: false };
  }
  try {
    const raw = window.localStorage.getItem(SETTINGS_KEY);
    if (!raw) {
      return { volume: DEFAULT_VOLUME, muted: false, repeatMode: 'off', shuffled: false };
    }
    const parsed = JSON.parse(raw) as {
      volume?: number;
      muted?: boolean;
      repeatMode?: RepeatMode;
      shuffled?: boolean;
    };
    const volume = Number.isFinite(parsed.volume)
      ? Math.min(100, Math.max(0, Number(parsed.volume)))
      : DEFAULT_VOLUME;
    const repeatMode: RepeatMode =
      parsed.repeatMode === 'all' || parsed.repeatMode === 'one' ? parsed.repeatMode : 'off';
    return {
      volume,
      muted: Boolean(parsed.muted),
      repeatMode,
      shuffled: Boolean(parsed.shuffled),
    };
  } catch {
    return { volume: DEFAULT_VOLUME, muted: false, repeatMode: 'off', shuffled: false };
  }
}

const persisted = readPersistedSettings();

export const player = $state({
  viewMode: 'normal' as ViewMode,
  currentSong: {
    id: 'init-galaxy',
    title: '在银河中孤独摇摆',
    artist: '知更鸟 / HOYO-MiX',
    album: '未知专辑',
    duration: '—',
    coverUrl: defaultCover,
  } as PlayerTrack,
  isPlaying: false,
  queue: [] as PlayerTrack[],
  isShuffled: persisted.shuffled,
  repeatMode: persisted.repeatMode as RepeatMode,
  volume: persisted.volume,
  isMuted: persisted.muted,
});

function persistSettings() {
  if (typeof window === 'undefined') {
    return;
  }
  window.localStorage.setItem(
    SETTINGS_KEY,
    JSON.stringify({
      volume: player.volume,
      muted: player.isMuted,
      repeatMode: player.repeatMode,
      shuffled: player.isShuffled,
    }),
  );
}

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

export function setQueue(tracks: PlayerTrack[]) {
  player.queue = tracks;
}

export function togglePlayByTrack(track: PlayerTrack) {
  console.info('[播放器状态] 调用按歌曲切换播放', {
    id: track.id,
    title: track.title,
    artist: track.artist,
    album: track.album,
    playback: track.playback,
  });
  if (sameTrack(player.currentSong, track)) {
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

export function playNextTrack(fromEnded = false): boolean {
  const queue = player.queue;
  if (queue.length === 0) {
    return false;
  }
  const currentIndex = Math.max(
    0,
    queue.findIndex((item) => sameTrack(item, player.currentSong)),
  );
  let nextIndex = currentIndex + 1;

  if (player.isShuffled && queue.length > 1) {
    do {
      nextIndex = Math.floor(Math.random() * queue.length);
    } while (nextIndex === currentIndex);
  } else if (nextIndex >= queue.length) {
    if (player.repeatMode === 'all' || !fromEnded) {
      nextIndex = 0;
    } else {
      return false;
    }
  }

  player.currentSong = queue[nextIndex];
  player.isPlaying = true;
  return true;
}

export function playPreviousTrack(): boolean {
  const queue = player.queue;
  if (queue.length === 0) {
    return false;
  }
  const currentIndex = Math.max(
    0,
    queue.findIndex((item) => sameTrack(item, player.currentSong)),
  );
  let prevIndex = currentIndex - 1;

  if (player.isShuffled && queue.length > 1) {
    do {
      prevIndex = Math.floor(Math.random() * queue.length);
    } while (prevIndex === currentIndex);
  } else if (prevIndex < 0) {
    prevIndex = queue.length - 1;
  }

  player.currentSong = queue[prevIndex];
  player.isPlaying = true;
  return true;
}

export function toggleShuffleMode() {
  player.isShuffled = !player.isShuffled;
  persistSettings();
}

export function cycleRepeatMode() {
  const modes: RepeatMode[] = ['off', 'all', 'one'];
  const current = modes.indexOf(player.repeatMode);
  player.repeatMode = modes[(current + 1) % modes.length];
  persistSettings();
}

export function setPlayerVolume(volume: number, options?: { persist?: boolean }) {
  player.volume = Math.min(100, Math.max(0, Math.round(volume)));
  if (player.volume > 0) {
    player.isMuted = false;
  }
  if (options?.persist !== false) {
    persistSettings();
  }
}

export function togglePlayerMuted() {
  player.isMuted = !player.isMuted;
  persistSettings();
}
