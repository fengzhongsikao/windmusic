import { writable } from 'svelte/store';
import { GetLocalAudioStream, GetMusicURL } from '../../../wailsjs/go/main/App';
import { localPathFromMetaJson, LOCAL_PLATFORM } from '@/lib/localMusic';
import { player, playNextTrack, setPlaying, type PlayerTrack } from '@/stores/playback/player.svelte';
import { loadLyricsForTrack, trackPlaybackKey } from '@/stores/playback/lyrics';
import { error as errorToast } from '@/stores/ui/toast';
import { recordRecentPlay } from '@/lib/wailsPlayer';

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
let switchingTrack = false;

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
    if (!audio.currentSrc && !audio.src) {
      audioError.set('播放歌曲失败');
      errorToast('播放歌曲失败');
      syncingFromAudio = true;
      setPlaying(false);
      syncingFromAudio = false;
      return;
    }
    void audio.play()
      .then(() => {
        syncingFromAudio = true;
        setPlaying(true);
        syncingFromAudio = false;
      })
      .catch(() => {
        audioError.set('播放歌曲失败');
        errorToast('播放歌曲失败');
        syncingFromAudio = true;
        setPlaying(false);
        syncingFromAudio = false;
      });
  } else if (!shouldPlay && !audio.paused) {
    audio.pause();
  }
}

async function loadAudioForTrack(track: PlayerTrack, shouldAutoPlay: boolean) {
  const token = ++audioLoadToken;
  switchingTrack = true;
  audioLoading.set(true);
  audioError.set('');
  audioReady.set(false);

  if (!audio) {
    switchingTrack = false;
    audioLoading.set(false);
    return;
  }

  const ctx = track.playback;
  let localPath = ctx?.localPath?.trim() ?? '';
  if (!localPath && ctx?.platform === LOCAL_PLATFORM) {
    localPath = localPathFromMetaJson(ctx.metaJson) || String(track.id ?? '').trim();
  }
  const isLocal = localPath !== '';
  if (!isLocal && (!ctx?.sourceId || !ctx.platform || !ctx.metaJson)) {
    syncingFromAudio = true;
    audio.pause();
    audio.removeAttribute('src');
    audio.load();
    syncingFromAudio = false;
    audioCurrentTime.set(0);
    audioDuration.set(0);
    if (shouldAutoPlay) {
      audioError.set('播放歌曲失败');
      errorToast('播放歌曲失败');
      syncingFromAudio = true;
      setPlaying(false);
      syncingFromAudio = false;
    }
    switchingTrack = false;
    audioLoading.set(false);
    return;
  }

  try {
    const url = isLocal
      ? await GetLocalAudioStream(localPath)
      : await GetMusicURL(ctx!.sourceId, ctx!.platform, '', ctx!.metaJson);
    if (token !== audioLoadToken || !audio) {
      return;
    }

    console.info('前端[音频引擎] 获取播放地址结果', {
      title: track.title,
      artist: track.artist,
      platform: isLocal ? 'local' : ctx!.platform,
      sourceId: isLocal ? 'local' : ctx!.sourceId,
      local: isLocal,
    });

    const playableUrl = url.trim();
    if (!playableUrl) {
      syncingFromAudio = true;
      audio.pause();
      audio.removeAttribute('src');
      audio.load();
      syncingFromAudio = false;
      audioError.set('播放歌曲失败');
      errorToast('播放歌曲失败');
      syncingFromAudio = true;
      setPlaying(false);
      syncingFromAudio = false;
      return;
    }

    syncingFromAudio = true;
    audio.pause();
    if (audio.src !== playableUrl) {
      audio.src = playableUrl;
      audio.load();
    }
    syncingFromAudio = false;

    if (shouldAutoPlay) {
      await audio.play();
      syncingFromAudio = true;
      setPlaying(true);
      syncingFromAudio = false;
      void recordRecentPlay(track);
    }
  } catch (err) {
    if (token !== audioLoadToken) {
      return;
    }
    syncingFromAudio = true;
    audio.pause();
    audio.removeAttribute('src');
    audio.load();
    syncingFromAudio = false;
    audioError.set('播放歌曲失败');
    errorToast('播放歌曲失败');
    syncingFromAudio = true;
    setPlaying(false);
    syncingFromAudio = false;
  } finally {
    if (token === audioLoadToken) {
      switchingTrack = false;
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
    if (syncingFromAudio || switchingTrack) {
      return;
    }
    if (!lastPlaying) {
      setPlaying(true);
    }
  });

  el.addEventListener('pause', () => {
    if (syncingFromAudio || switchingTrack) {
      return;
    }
    if (lastPlaying) {
      setPlaying(false);
    }
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
      if (playing) {
        syncingFromAudio = true;
        setPlaying(false);
        syncingFromAudio = false;
        void loadAudioForTrack(track, true);
      } else {
        syncPlayState(false);
      }
    }
    return;
  }

  if (audio) {
    const ctx = track.playback;
    const hasPlayback = Boolean(
      ctx?.localPath?.trim() || (ctx?.sourceId && ctx.platform && ctx.metaJson),
    );
    const hasSource = Boolean(audio.currentSrc || audio.src);
    if (playing && hasPlayback && !hasSource) {
      void loadAudioForTrack(track, true);
      return;
    }
    syncPlayState(playing);
  }
}

export function initAudioEngine(el: HTMLAudioElement, root: HTMLElement) {
  audio = el;
  audioRoot = root;
  attachAudioListeners(el);
  setAudioVolume(player.volume, player.isMuted);
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
