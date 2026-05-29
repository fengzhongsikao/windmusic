import type { TrackItem } from '@/lib/track';
import { createDefaultPlayerTrack } from '@/lib/defaultPlayerTrack';

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

let shuffleOrder: number[] = [];

function fisherYatesOrder(length: number): number[] {
  const order = Array.from({ length }, (_, index) => index);
  for (let i = length - 1; i > 0; i -= 1) {
    const j = Math.floor(Math.random() * (i + 1));
    [order[i], order[j]] = [order[j], order[i]];
  }
  return order;
}

function rebuildShuffleOrder() {
  const { queue } = player;
  if (queue.length === 0) {
    shuffleOrder = [];
    return;
  }
  shuffleOrder = fisherYatesOrder(queue.length);
  const currentIndex = queue.findIndex((item) => sameTrack(item, player.currentSong));
  if (currentIndex >= 0) {
    const pos = shuffleOrder.indexOf(currentIndex);
    if (pos > 0) {
      shuffleOrder[pos] = shuffleOrder[0];
      shuffleOrder[0] = currentIndex;
    }
  }
}

function syncShuffleOrderAfterQueueChange() {
  if (!player.isShuffled) {
    shuffleOrder = [];
    return;
  }
  if (player.queue.length === 0) {
    shuffleOrder = [];
    return;
  }
  if (shuffleOrder.length !== player.queue.length) {
    rebuildShuffleOrder();
  }
}

export function isCurrentTrack(track: PlayerTrack): boolean {
  return sameTrack(track, player.currentSong);
}

export function repeatModeLabel(mode: RepeatMode): string {
  if (mode === 'one') return '单曲循环';
  if (mode === 'all') return '列表循环';
  return '顺序播放';
}

const defaultTrack = createDefaultPlayerTrack();

export const player = $state({
  viewMode: 'normal' as ViewMode,
  currentSong: defaultTrack,
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
  syncShuffleOrderAfterQueueChange();
}

export function playQueueTrack(index: number) {
  const track = player.queue[index];
  if (!track) {
    return;
  }
  player.currentSong = track;
  player.isPlaying = true;
}

export function removeQueueTrack(index: number) {
  if (index < 0 || index >= player.queue.length) {
    return;
  }
  const removingCurrent = sameTrack(player.queue[index], player.currentSong);
  player.queue = player.queue.filter((_, i) => i !== index);
  if (player.isShuffled) {
    shuffleOrder = shuffleOrder
      .filter((i) => i !== index)
      .map((i) => (i > index ? i - 1 : i));
    syncShuffleOrderAfterQueueChange();
  }
  if (removingCurrent) {
    if (player.queue.length === 0) {
      player.isPlaying = false;
      return;
    }
    playNextTrack();
  }
}

export function clearQueue() {
  player.queue = [];
  shuffleOrder = [];
}

export function togglePlayByTrack(track: PlayerTrack) {
  const queueIndex = player.queue.findIndex((item) => sameTrack(item, track));

  if (sameTrack(player.currentSong, track)) {
    player.isPlaying = !player.isPlaying;
    if (queueIndex < 0) {
      player.queue = [...player.queue, track];
      syncShuffleOrderAfterQueueChange();
    }
    return;
  }

  if (queueIndex >= 0) {
    player.currentSong = player.queue[queueIndex];
    player.isPlaying = true;
    return;
  }

  player.queue = [...player.queue, track];
  syncShuffleOrderAfterQueueChange();
  player.currentSong = track;
  player.isPlaying = true;
}

export function playAllTracks(tracks: PlayerTrack[]) {
  if (tracks.length === 0) {
    return;
  }
  setQueue(tracks);
  player.currentSong = tracks[0];
  player.isPlaying = true;
}

export function togglePlayerPlayback() {
  player.isPlaying = !player.isPlaying;
}

export function setPlaying(playing: boolean) {
  player.isPlaying = playing;
}

function resolveNextQueueIndex(fromEnded: boolean): number | null {
  const queue = player.queue;
  if (queue.length === 0) {
    return null;
  }
  if (queue.length === 1) {
    return 0;
  }

  const currentIndex = queue.findIndex((item) => sameTrack(item, player.currentSong));

  if (player.isShuffled && shuffleOrder.length === queue.length) {
    const orderPos = currentIndex < 0 ? -1 : shuffleOrder.indexOf(currentIndex);
    let nextOrderPos = orderPos + 1;
    if (nextOrderPos >= shuffleOrder.length) {
      if (player.repeatMode === 'all' || !fromEnded) {
        nextOrderPos = 0;
      } else {
        return null;
      }
    }
    return shuffleOrder[nextOrderPos];
  }

  let nextIndex = currentIndex < 0 ? 0 : currentIndex + 1;
  if (nextIndex >= queue.length) {
    if (player.repeatMode === 'all' || !fromEnded) {
      nextIndex = 0;
    } else {
      return null;
    }
  }
  return nextIndex;
}

function resolvePreviousQueueIndex(): number | null {
  const queue = player.queue;
  if (queue.length === 0) {
    return null;
  }
  if (queue.length === 1) {
    return 0;
  }

  const currentIndex = queue.findIndex((item) => sameTrack(item, player.currentSong));

  if (player.isShuffled && shuffleOrder.length === queue.length) {
    const orderPos = currentIndex < 0 ? 0 : shuffleOrder.indexOf(currentIndex);
    const prevOrderPos = orderPos <= 0 ? shuffleOrder.length - 1 : orderPos - 1;
    return shuffleOrder[prevOrderPos];
  }

  if (currentIndex <= 0) {
    return queue.length - 1;
  }
  return currentIndex - 1;
}

export function playNextTrack(fromEnded = false): boolean {
  const nextIndex = resolveNextQueueIndex(fromEnded);
  if (nextIndex == null) {
    return false;
  }
  player.currentSong = player.queue[nextIndex];
  player.isPlaying = true;
  return true;
}

export function playPreviousTrack(): boolean {
  const prevIndex = resolvePreviousQueueIndex();
  if (prevIndex == null) {
    return false;
  }
  player.currentSong = player.queue[prevIndex];
  player.isPlaying = true;
  return true;
}

export function toggleShuffleMode() {
  player.isShuffled = !player.isShuffled;
  if (player.isShuffled) {
    rebuildShuffleOrder();
  } else {
    shuffleOrder = [];
  }
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
