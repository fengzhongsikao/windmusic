import { writable } from 'svelte/store';
import { GetMusicURL } from '../../wailsjs/go/main/App';
import { player, playNextTrack, setPlaying, type PlayerTrack } from '@/stores/player.svelte';
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
let lastPlaying = false;
let syncingFromAudio = false;

function errorMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}

function resolveMusicUrl(raw: string): string {
  const text = raw?.trim() ?? '';
  if (!text) {
    return '';
  }
  if (/^https?:\/\//i.test(text)) {
    return text;
  }

  try {
    const parsed = JSON.parse(text) as
      | string
      | {
      code?: number;
      msg?: string;
      data?: unknown;
      url?: unknown;
    };
    if (typeof parsed === 'string' && parsed.trim()) {
      return parsed.trim();
    }
    if (typeof parsed === 'object' && parsed !== null) {
      if (typeof parsed.data === 'string' && parsed.data.trim()) {
        return parsed.data.trim();
      }
      if (typeof parsed.url === 'string' && parsed.url.trim()) {
        return parsed.url.trim();
      }
      if (parsed.code !== undefined && parsed.code !== 0) {
        throw new Error(parsed.msg || `音源返回错误 code=${parsed.code}`);
      }
    }
  } catch {
    return text;
  }

  return '';
}

function normalizePlayableUrl(url: string): string {
  const text = url.trim();
  if (!text) {
    return '';
  }
  // 在 WebView/浏览器环境中，http 音频有时会被拦截；优先升级为 https。
  if (text.startsWith('http://')) {
    return `https://${text.slice('http://'.length)}`;
  }
  return text;
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
      setPlaying(false);
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

    console.info('前端[音频引擎] 获取播放地址结果', {
      title: track.title,
      artist: track.artist,
      platform: ctx.platform,
      sourceId: ctx.sourceId,
      musicUrl: url,
    });

    const resolved = resolveMusicUrl(url);
    const playableUrl = normalizePlayableUrl(resolved);
    if (!playableUrl) {
      audio.removeAttribute('src');
      audio.load();
      audioError.set('未获取到播放地址');
      return;
    }

    if (audio.src !== playableUrl) {
      audio.src = playableUrl;
      audio.load();
    }

    const shouldPlay = lastPlaying;
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
  el.addEventListener('canplay', () => {
    audioError.set('');
  });
  el.addEventListener('error', () => {
    const mediaError = el.error;
    if (!mediaError) {
      audioError.set('音频加载失败');
      return;
    }
    const codeLabel =
      mediaError.code === MediaError.MEDIA_ERR_ABORTED
        ? '播放被中止'
        : mediaError.code === MediaError.MEDIA_ERR_NETWORK
          ? '网络错误'
          : mediaError.code === MediaError.MEDIA_ERR_DECODE
            ? '音频解码失败'
            : mediaError.code === MediaError.MEDIA_ERR_SRC_NOT_SUPPORTED
              ? '音频地址不可用'
              : '未知错误';
    const message = `音频播放失败：${codeLabel}`;
    console.error('[音频引擎] audio 元素错误', {
      code: mediaError.code,
      message: mediaError.message,
      src: el.currentSrc || el.src,
    });
    audioError.set(message);
    syncingFromAudio = true;
    setPlaying(false);
    syncingFromAudio = false;
  });

  el.addEventListener('play', () => {
    syncingFromAudio = true;
    if (!lastPlaying) {
      setPlaying(true);
    }
    syncingFromAudio = false;
  });

  el.addEventListener('pause', () => {
    syncingFromAudio = true;
    if (lastPlaying) {
      setPlaying(false);
    }
    syncingFromAudio = false;
  });

  el.addEventListener('ended', () => {
    if (player.repeatMode === 'one') {
      el.currentTime = 0;
      void el.play().catch(() => {
        syncingFromAudio = true;
        setPlaying(false);
        syncingFromAudio = false;
      });
      return;
    }
    if (playNextTrack(true)) {
      return;
    }
    syncingFromAudio = true;
    setPlaying(false);
    syncingFromAudio = false;
  });
}

function onTrackOrPlayStateChange(track: PlayerTrack, playing: boolean) {
  const key = trackPlaybackKey(track);
  lastPlaying = playing;

  if (key !== lastTrackKey) {
    lastTrackKey = key;
    void loadLyricsForTrack(track);
    if (audio) {
      void loadAudioForTrack(track);
    }
  }

  if (audio) {
    syncPlayState(playing);
  }
}

export function initAudioEngine(el: HTMLAudioElement, root: HTMLElement) {
  audio = el;
  audioRoot = root;
  attachAudioListeners(el);
}

export function syncPlayerState(track: PlayerTrack, playing: boolean) {
  onTrackOrPlayStateChange(track, playing);
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
