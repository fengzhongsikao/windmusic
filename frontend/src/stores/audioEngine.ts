import { writable, get } from 'svelte/store';
import { GetMusicURL } from '../../wailsjs/go/main/App';
import { playerState, type PlayerTrack } from '@/stores/player';
import { loadLyricsForTrack, trackPlaybackKey } from '@/stores/lyrics';

export const audioCurrentTime = writable(0);
export const audioDuration = writable(0);
export const audioLoading = writable(false);
export const audioError = writable('');
export const audioReady = writable(false);

let audio: HTMLAudioElement | null = null;
let audioRoot: HTMLElement | null = null;
let mountCount = 0;
let listenersAttached = false;

let audioLoadToken = 0;
let lastTrackKey = '';
let syncingFromAudio = false;

function errorMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}

function updateTimeState() {
  if (!audio) {
    return;
  }
  audioCurrentTime.set(audio.currentTime);
  const duration = Number.isFinite(audio.duration) ? audio.duration : 0;
  audioDuration.set(duration);
}

function syncPlayState(shouldPlay: boolean) {
  if (!audio || syncingFromAudio) {
    return;
  }
  if (shouldPlay && audio.paused) {
    void audio.play().catch(() => {
      playerState.update((state) => ({ ...state, isPlaying: false }));
    });
  } else if (!shouldPlay && !audio.paused) {
    audio.pause();
  }
}

async function loadAudioForTrack(track: PlayerTrack) {
  const token = ++audioLoadToken;
  audioLoading.set(true);
  audioError.set('');
  audioReady.set(false);

  if (!audio) {
    audioLoading.set(false);
    return;
  }

  const ctx = track.playback;
  if (!ctx?.sourceId || !ctx.platform || !ctx.metaJson) {
    audio.removeAttribute('src');
    audio.load();
    audioCurrentTime.set(0);
    audioDuration.set(0);
    audioLoading.set(false);
    return;
  }

  try {
    const url = await GetMusicURL(ctx.sourceId, ctx.platform, '', ctx.metaJson);
    if (token !== audioLoadToken || !audio) {
      return;
    }

    const trimmed = url?.trim();
    if (!trimmed) {
      audio.removeAttribute('src');
      audio.load();
      audioError.set('未获取到播放地址');
      return;
    }

    if (audio.src !== trimmed) {
      audio.src = trimmed;
      audio.load();
    }

    const shouldPlay = get(playerState).isPlaying;
    if (shouldPlay) {
      await audio.play();
    }
  } catch (err) {
    if (token !== audioLoadToken) {
      return;
    }
    audio.removeAttribute('src');
    audio.load();
    audioError.set(errorMessage(err));
  } finally {
    if (token === audioLoadToken) {
      audioLoading.set(false);
    }
  }
}

function attachAudioListeners(el: HTMLAudioElement) {
  if (listenersAttached) {
    return;
  }
  listenersAttached = true;

  el.addEventListener('timeupdate', updateTimeState);
  el.addEventListener('loadedmetadata', () => {
    updateTimeState();
    audioReady.set(true);
  });
  el.addEventListener('durationchange', updateTimeState);

  el.addEventListener('play', () => {
    syncingFromAudio = true;
    playerState.update((state) => (state.isPlaying ? state : { ...state, isPlaying: true }));
    syncingFromAudio = false;
  });

  el.addEventListener('pause', () => {
    syncingFromAudio = true;
    playerState.update((state) => (!state.isPlaying ? state : { ...state, isPlaying: false }));
    syncingFromAudio = false;
  });

  el.addEventListener('ended', () => {
    syncingFromAudio = true;
    playerState.update((state) => ({ ...state, isPlaying: false }));
    syncingFromAudio = false;
  });
}

function onTrackOrPlayStateChange() {
  const state = get(playerState);
  const key = trackPlaybackKey(state.currentTrack);

  if (key !== lastTrackKey) {
    lastTrackKey = key;
    void loadLyricsForTrack(state.currentTrack);
    if (audio) {
      void loadAudioForTrack(state.currentTrack);
    }
  }

  if (audio) {
    syncPlayState(state.isPlaying);
  }
}

export function initAudioEngine(el: HTMLAudioElement, root: HTMLElement) {
  audio = el;
  audioRoot = root;
  attachAudioListeners(el);
  onTrackOrPlayStateChange();
}

export function startAudioSync() {
  playerState.subscribe(() => {
    onTrackOrPlayStateChange();
  });
}

export function getAudioElement() {
  return audio;
}

export function getAudioRoot() {
  return audioRoot;
}

export function seekAudio(seconds: number) {
  if (!audio) {
    return;
  }
  const clamped = Math.max(0, Number.isFinite(audio.duration) ? Math.min(seconds, audio.duration) : seconds);
  audio.currentTime = clamped;
  audioCurrentTime.set(clamped);
}

export function setAudioVolume(percent: number, muted: boolean) {
  if (!audio) {
    return;
  }
  const level = Math.min(100, Math.max(0, percent)) / 100;
  audio.muted = muted;
  audio.volume = muted ? 0 : level;
}

export function mountAudioTo(container: HTMLElement) {
  if (!audio || !audioRoot) {
    return () => {};
  }

  mountCount += 1;
  container.appendChild(audio);
  audio.controls = true;
  audio.className = 'w-full rounded-lg';

  return () => {
    mountCount = Math.max(0, mountCount - 1);
    if (mountCount === 0 && audio && audioRoot) {
      audio.controls = false;
      audio.className = '';
      audioRoot.appendChild(audio);
    }
  };
}
