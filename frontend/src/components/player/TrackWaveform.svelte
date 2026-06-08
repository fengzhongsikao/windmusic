<script lang="ts">
  import {
    audioCurrentTime,
    audioDuration,
    audioReady,
    getAudioSourceUrl,
    seekAudio,
  } from '@/stores/playback/audioEngine';
  import { player } from '@/stores/playback/player.svelte';
  import { playerUiSettings } from '@/stores/playback/playerSettings.svelte';
  import { loadWaveformPeaks, shouldFetchWaveformPeaks } from '@/lib/playback/waveformPeaks';
  import type { WaveformSpreadMode } from '@/lib/playback/waveformSpread';
  import {
    computeRestBarHeights,
    computeVisibleBarCount,
    drawLiveWaveform,
    samplePeaksBarHeights,
    sampleSpreadMotionBarHeights,
    smoothHeights,
    type WaveformColors,
  } from '@/lib/playback/waveform';

  interface Props {
    tone?: 'light' | 'dark';
    height?: number;
  }

  let { tone = 'light', height = 60 }: Props = $props();

  let canvasEl = $state<HTMLCanvasElement | null>(null);
  let viewportEl = $state<HTMLDivElement | null>(null);
  let viewportWidth = $state(360);
  let barCount = $derived(computeVisibleBarCount(viewportWidth));
  let barHeights = $state(new Float32Array(0));
  let targetHeights = $state(new Float32Array(0));
  let scrubTime = $state<number | null>(null);
  let syncedAudioTime = $state(0);
  let syncedAtMs = $state(0);
  let waveformPeaks = $state<Float32Array | null>(null);
  let peaksSourceUrl = $state('');
  let lastPaintRevision = $state(-1);
  let motionPhaseOriginMs = $state(0);

  let currentTime = $derived($audioCurrentTime);
  let duration = $derived($audioDuration);
  let displayTime = $derived(scrubTime ?? currentTime);
  let progressPercent = $derived(duration > 0 ? (displayTime / duration) * 100 : 0);
  let isPlaying = $derived(player.isPlaying && $audioReady);
  let canDriveWave = $derived(
    isPlaying && !player.isMuted && player.volume > 0,
  );
  let spreadMode = $derived(playerUiSettings.waveformSpread);
  let spreadRevision = $derived(playerUiSettings.waveformSpreadRevision);

  const colors = $derived<WaveformColors>(
    tone === 'light'
      ? {
          wave: 'rgba(255, 255, 255, 0.92)',
          progress: 'rgba(255, 255, 255, 0.98)',
        }
      : {
          wave: 'rgba(0, 0, 0, 0.55)',
          progress: 'rgba(0, 0, 0, 0.82)',
        },
  );

  function formatTime(seconds: number): string {
    const safe = Math.max(0, Number.isFinite(seconds) ? seconds : 0);
    const min = Math.floor(safe / 60);
    const sec = Math.floor(safe % 60);
    return `${min.toString().padStart(2, '0')}:${sec.toString().padStart(2, '0')}`;
  }

  function ensureBuffers(count: number) {
    if (barHeights.length === count && targetHeights.length === count) {
      return;
    }
    barHeights = new Float32Array(count);
    targetHeights = new Float32Array(count);
  }

  function sampleTimeForWaveform(): number {
    if (scrubTime !== null) {
      return scrubTime;
    }
    if (canDriveWave && duration > 0) {
      const elapsed = (performance.now() - syncedAtMs) / 1000;
      return Math.min(duration, Math.max(0, syncedAudioTime + elapsed));
    }
    return syncedAudioTime;
  }

  function paintWaveform(mode: WaveformSpreadMode) {
    const canvas = canvasEl;
    if (!canvas || barCount <= 0) return;

    ensureBuffers(barCount);

    let showingLive = false;
    const peaks = waveformPeaks;
    const time = sampleTimeForWaveform();
    const trackDuration = duration;
    const phaseMs = performance.now() - motionPhaseOriginMs;
    const waveOpts = { live: canDriveWave, phaseMs, spreadMode: mode };
    let sampled = false;

    if (peaks && trackDuration > 0) {
      sampled = samplePeaksBarHeights(peaks, trackDuration, time, barCount, targetHeights, waveOpts);
    } else if (canDriveWave) {
      sampled = sampleSpreadMotionBarHeights(barCount, targetHeights, waveOpts);
    }

    if (sampled) {
      showingLive = canDriveWave;
      const modeChanged = spreadRevision !== lastPaintRevision;
      if (modeChanged) {
        lastPaintRevision = spreadRevision;
      }
      const smoothFactor = modeChanged ? 1 : canDriveWave ? 0.42 : 0.28;
      smoothHeights(barHeights, targetHeights, smoothFactor);
    } else {
      computeRestBarHeights(barCount, targetHeights);
      smoothHeights(barHeights, targetHeights, 0.2);
    }

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    drawLiveWaveform({
      ctx,
      heights: barHeights,
      viewportWidth,
      height,
      colors,
      active: showingLive,
      devicePixelRatio: window.devicePixelRatio,
    });
  }

  function seekFromClientX(clientX: number, trackEl: HTMLElement) {
    if (duration <= 0) return;
    const rect = trackEl.getBoundingClientRect();
    if (rect.width <= 0) return;
    const ratio = Math.min(1, Math.max(0, (clientX - rect.left) / rect.width));
    seekAudio(ratio * duration);
  }

  function handleProgressPointerDown(event: PointerEvent) {
    const trackEl = event.currentTarget as HTMLElement;
    trackEl.setPointerCapture(event.pointerId);
    scrubTime = displayTime;
    seekFromClientX(event.clientX, trackEl);

    const onMove = (moveEvent: PointerEvent) => {
      const rect = trackEl.getBoundingClientRect();
      if (rect.width <= 0 || duration <= 0) return;
      const ratio = Math.min(1, Math.max(0, (moveEvent.clientX - rect.left) / rect.width));
      scrubTime = ratio * duration;
    };

    const onUp = (upEvent: PointerEvent) => {
      seekFromClientX(upEvent.clientX, trackEl);
      scrubTime = null;
      trackEl.releasePointerCapture(upEvent.pointerId);
      trackEl.removeEventListener('pointermove', onMove);
      trackEl.removeEventListener('pointerup', onUp);
      trackEl.removeEventListener('pointercancel', onUp);
    };

    trackEl.addEventListener('pointermove', onMove);
    trackEl.addEventListener('pointerup', onUp);
    trackEl.addEventListener('pointercancel', onUp);
  }

  $effect(() => {
    const viewport = viewportEl;
    if (!viewport) return;

    const updateWidth = () => {
      viewportWidth = Math.max(1, viewport.clientWidth);
    };

    updateWidth();
    const observer = new ResizeObserver(updateWidth);
    observer.observe(viewport);
    return () => observer.disconnect();
  });

  $effect(() => {
    syncedAudioTime = currentTime;
    syncedAtMs = performance.now();
  });

  $effect(() => {
    const sourceUrl = getAudioSourceUrl();
    if (!sourceUrl || sourceUrl === peaksSourceUrl) {
      return;
    }

    peaksSourceUrl = sourceUrl;
    waveformPeaks = null;

    if (!shouldFetchWaveformPeaks(sourceUrl)) {
      return;
    }

    void loadWaveformPeaks(sourceUrl).then((peaks) => {
      if (peaksSourceUrl === sourceUrl) {
        waveformPeaks = peaks;
      }
    });
  });

  $effect(() => {
    spreadRevision;
    motionPhaseOriginMs = performance.now();
  });

  $effect(() => {
    const mode = spreadMode;
    spreadRevision;
    canDriveWave;
    waveformPeaks;
    displayTime;
    barCount;
    viewportWidth;
    height;
    colors;
    duration;

    let frameId = 0;
    let frames = 0;
    const decayFrames = 24;

    const loop = () => {
      paintWaveform(mode);
      frames += 1;
      if (!canDriveWave && frames >= decayFrames) {
        return;
      }
      frameId = requestAnimationFrame(loop);
    };

    frames = 0;
    frameId = requestAnimationFrame(loop);
    return () => cancelAnimationFrame(frameId);
  });
</script>

<div class="track-waveform">
  <div class="waveform-wrap" aria-hidden="true">
    <div bind:this={viewportEl} class="waveform-viewport">
      <canvas bind:this={canvasEl} class="waveform-canvas" width={viewportWidth} {height}></canvas>
    </div>
  </div>

  <div class="track-progress">
    <span class="time">{formatTime(displayTime)}</span>
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      class="progress-track"
      role="slider"
      tabindex="0"
      aria-label="播放进度"
      aria-valuemin={0}
      aria-valuemax={Math.max(duration, 0)}
      aria-valuenow={displayTime}
      aria-disabled={duration <= 0}
      onpointerdown={handleProgressPointerDown}
      onkeydown={(event) => {
        if (duration <= 0) return;
        const step = event.shiftKey ? 10 : 5;
        if (event.key === 'ArrowLeft') {
          event.preventDefault();
          seekAudio(Math.max(0, displayTime - step));
        } else if (event.key === 'ArrowRight') {
          event.preventDefault();
          seekAudio(Math.min(duration, displayTime + step));
        }
      }}
    >
      <div class="progress-line">
        <div class="progress-fill" style:width="{progressPercent}%"></div>
      </div>
    </div>
    <span class="time">{formatTime(duration)}</span>
  </div>
</div>

<style>
  .track-waveform {
    width: 100%;
    max-width: 360px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .waveform-wrap {
    position: relative;
    width: 100%;
    height: 60px;
    mask-image: linear-gradient(
      90deg,
      transparent 0%,
      #000 8%,
      #000 92%,
      transparent 100%
    );
  }

  .waveform-viewport {
    width: 100%;
    height: 100%;
    overflow: hidden;
  }

  .waveform-canvas {
    display: block;
    width: 100%;
    height: 100%;
    pointer-events: none;
  }

  .track-progress {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
    padding: 0 2px;
  }

  .time {
    flex-shrink: 0;
    min-width: 38px;
    font-size: 11px;
    line-height: 1;
    color: rgba(255, 255, 255, 0.42);
    font-variant-numeric: tabular-nums;
  }

  .time:first-child {
    text-align: left;
  }

  .time:last-child {
    text-align: right;
  }

  .progress-track {
    flex: 1;
    min-width: 0;
    height: 12px;
    display: flex;
    align-items: center;
    cursor: pointer;
    touch-action: none;
  }

  .progress-line {
    width: 100%;
    height: 2px;
    border-radius: 1px;
    background: rgba(255, 255, 255, 0.22);
    position: relative;
    overflow: hidden;
  }

  .progress-fill {
    position: absolute;
    top: 0;
    left: 0;
    height: 100%;
    border-radius: inherit;
    background: rgba(255, 255, 255, 0.92);
    pointer-events: none;
  }

  .progress-track:focus-visible {
    outline: 1px solid rgba(255, 255, 255, 0.35);
    outline-offset: 3px;
    border-radius: 2px;
  }
</style>
