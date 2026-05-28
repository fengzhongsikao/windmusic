<!--
  参考稿风格：毛玻璃暖色背景、左视觉右歌词、底部控制 dock。
-->
<script lang="ts">
  import {
    Music,
    Play,
    Pause,
    Volume2,
    Volume1,
    Volume,
    VolumeX,
    Shuffle,
    SkipBack,
    SkipForward,
    Share2,
    Heart,
    MicVocal,
    Settings,
    ListMusic,
    Star,
    LoaderCircle,
  } from '@lucide/svelte';
  import SpectrumAnalyzer from '@/components/SpectrumAnalyzer.svelte';
  import { playerState, togglePlayerPlayback } from '@/stores/player';
  import { closeSongDetailDrawer } from '@/stores/songDetailDrawer';
  import { parseLrc, findActiveLineIndex, type LrcLine } from '@/lib/lrc';
  import { lrcRaw, lyricLoading, lyricError } from '@/stores/lyrics';
  import {
    audioCurrentTime,
    audioDuration,
    audioLoading,
    audioError,
    mountAudioTo,
    seekAudio,
    setAudioVolume,
  } from '@/stores/audioEngine';
  import defaultCover from '@/assets/images/default.jpg';

  let track = $derived($playerState.currentTrack);
  let coverSrc = $derived(track.coverUrl?.trim() || defaultCover);
  let isPlaying = $derived($playerState.isPlaying);
  let currentTime = $derived($audioCurrentTime);
  let duration = $derived($audioDuration);

  let volume = $state(75);
  let isMuted = $state(false);

  let audioHostEl = $state<HTMLDivElement | null>(null);
  let lyricsContainerEl = $state<HTMLDivElement | null>(null);
  let lineElements = $state<Record<number, HTMLButtonElement | undefined>>({});

  let lyricLines = $derived(parseLrc($lrcRaw));
  let activeIndex = $derived(findActiveLineIndex(lyricLines, currentTime));
  let progressPercent = $derived(duration > 0 ? (currentTime / duration) * 100 : 0);

  let artistLine = $derived(
    [track.artist, track.album].filter(Boolean).join(' - ') || track.artist,
  );

  function seekToLine(index: number) {
    const line = lyricLines[index];
    if (!line) return;
    seekAudio(line.time);
  }

  function lineLabel(line: LrcLine, index: number): string {
    const prefix = index === activeIndex ? '正在播放：' : '';
    return `${prefix}${line.text || '…'}`;
  }

  function formatTime(seconds: number): string {
    const min = Math.floor(seconds / 60);
    const sec = Math.floor(seconds % 60);
    return `${min.toString().padStart(2, '0')}:${sec.toString().padStart(2, '0')}`;
  }

  function togglePlay() {
    togglePlayerPlayback();
  }

  function toggleMute() {
    isMuted = !isMuted;
  }

  function handleProgressClick(e: MouseEvent) {
    if (duration <= 0) return;
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    seekAudio(((e.clientX - rect.left) / rect.width) * duration);
  }

  function handleProgressKeydown(e: KeyboardEvent) {
    if (duration <= 0) return;
    const step = 5;
    switch (e.key) {
      case 'ArrowLeft':
      case 'ArrowDown':
        e.preventDefault();
        seekAudio(Math.max(0, currentTime - step));
        break;
      case 'ArrowRight':
      case 'ArrowUp':
        e.preventDefault();
        seekAudio(Math.min(duration, currentTime + step));
        break;
      case 'Home':
        e.preventDefault();
        seekAudio(0);
        break;
      case 'End':
        e.preventDefault();
        seekAudio(duration);
        break;
    }
  }

  function handleVolumeClick(e: MouseEvent) {
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    volume = Math.round(((e.clientX - rect.left) / rect.width) * 100);
    isMuted = false;
  }

  function handleVolumeKeydown(e: KeyboardEvent) {
    const step = 5;
    switch (e.key) {
      case 'ArrowLeft':
      case 'ArrowDown':
        e.preventDefault();
        isMuted = false;
        volume = Math.max(0, volume - step);
        break;
      case 'ArrowRight':
      case 'ArrowUp':
        e.preventDefault();
        isMuted = false;
        volume = Math.min(100, volume + step);
        break;
    }
  }

  $effect(() => {
    const host = audioHostEl;
    if (!host) return;
    return mountAudioTo(host);
  });

  $effect(() => {
    setAudioVolume(volume, isMuted);
  });

  $effect(() => {
    const index = activeIndex;
    const container = lyricsContainerEl;
    const lineEl = index >= 0 ? lineElements[index] : undefined;
    if (!container || !lineEl || index < 0) return;

    const scrollTop =
      lineEl.offsetTop - container.clientHeight / 2 + lineEl.offsetHeight / 2;
    container.scrollTo({ top: Math.max(0, scrollTop), behavior: 'smooth' });
  });
</script>

<div class="player-view">
  <div bind:this={audioHostEl} class="audio-host" aria-hidden="true"></div>

  <div class="player-stage">
    <aside class="visual-side" aria-labelledby="track-heading">
      <div class="cover-wrap">
        <div class="cover-ring" aria-hidden="true"></div>
        <div class="cover-disc" class:spinning={isPlaying}>
          <img src={coverSrc} alt={`${track.title} 封面`} class="cover-img" />
        </div>
      </div>

      <div class="track-meta">
        <h2 id="track-heading" class="track-title">{track.title}</h2>
        <p class="track-sub">{artistLine}</p>
      </div>

      <SpectrumAnalyzer active={isPlaying} tone="light" bars={44} />

      <div class="mini-progress">
        <span class="mini-time">{formatTime(currentTime)}</span>
        <div
          class="mini-track"
          role="slider"
          tabindex="0"
          aria-label="播放进度"
          aria-valuemin={0}
          aria-valuemax={duration}
          aria-valuenow={currentTime}
          onclick={handleProgressClick}
          onkeydown={handleProgressKeydown}
        >
          <div class="mini-fill" style="width: {progressPercent}%"></div>
        </div>
        <span class="mini-time">{formatTime(duration)}</span>
      </div>
    </aside>

    <section class="lyrics-side" aria-label="歌词">
      {#if $lyricLoading}
        <span class="lyric-loading" aria-hidden="true">
          <LoaderCircle size={18} />
        </span>
      {/if}
      {#if $lyricError || $audioError || $audioLoading}
        <p class="status-hint" role="status">
          {#if $audioLoading}正在加载音频…{/if}
          {#if $audioError}{$audioError}{/if}
          {#if $lyricError}{$lyricError}{/if}
        </p>
      {/if}

      {#if lyricLines.length === 0}
        <div class="lyrics-empty">
          <Music size={28} strokeWidth={1.25} aria-hidden="true" />
          <p>暂无歌词</p>
        </div>
      {:else}
        <div
          bind:this={lyricsContainerEl}
          class="lyrics-scroll"
          role="group"
          aria-live="polite"
        >
          {#each lyricLines as line, index (index)}
            <button
              type="button"
              bind:this={lineElements[index]}
              class="lyric-line"
              class:active={index === activeIndex}
              aria-current={index === activeIndex ? 'true' : undefined}
              aria-label={lineLabel(line, index)}
              onclick={() => seekToLine(index)}
            >
              {line.text || '…'}
            </button>
          {/each}
        </div>
      {/if}
    </section>
  </div>

  <footer class="bottom-dock">
    <div class="dock-left">
      <img src={coverSrc} alt="" class="dock-thumb" />
      <div class="dock-meta">
        <span class="dock-title">{track.title}</span>
        <span class="dock-artist">{artistLine}</span>
      </div>
    </div>

    <div class="dock-center">
      <div class="transport">
        <button type="button" class="dock-icon" aria-label="随机播放" title="随机播放">
          <Shuffle size={17} />
        </button>
        <button type="button" class="dock-icon" aria-label="上一首" title="上一首">
          <SkipBack size={20} />
        </button>
        <button
          type="button"
          class="dock-play"
          onclick={togglePlay}
          aria-label={isPlaying ? '暂停' : '播放'}
        >
          {#if isPlaying}
            <Pause size={22} />
          {:else}
            <Play size={22} />
          {/if}
        </button>
        <button type="button" class="dock-icon" aria-label="下一首" title="下一首">
          <SkipForward size={20} />
        </button>
        <button
          type="button"
          class="dock-icon"
          onclick={toggleMute}
          aria-label={isMuted ? '取消静音' : '音量'}
        >
          {#if isMuted}
            <VolumeX size={17} />
          {:else}
            <Volume2 size={17} />
          {/if}
        </button>
      </div>

      <div class="main-progress">
        <div
          class="main-track"
          role="slider"
          tabindex="0"
          aria-label="播放进度"
          aria-valuemin={0}
          aria-valuemax={duration}
          aria-valuenow={currentTime}
          onclick={handleProgressClick}
          onkeydown={handleProgressKeydown}
        >
          <div class="main-fill" style="width: {progressPercent}%">
            <span class="main-thumb"></span>
          </div>
        </div>
      </div>
    </div>

    <div class="dock-right">
      <div class="stars" aria-hidden="true">
        {#each [1, 2, 3, 4, 5] as n (n)}
          <Star size={14} class={n <= 3 ? 'star-on' : 'star-off'} />
        {/each}
      </div>
      <button type="button" class="dock-icon" aria-label="分享" title="分享">
        <Share2 size={17} />
      </button>
      <button type="button" class="dock-icon" aria-label="喜欢" title="喜欢">
        <Heart size={17} />
      </button>
      <button
        type="button"
        class="dock-icon"
        aria-label="关闭歌词"
        title="关闭"
        onclick={closeSongDetailDrawer}
      >
        <MicVocal size={17} />
      </button>
      <button type="button" class="dock-icon" aria-label="音效" title="音效">
        <Settings size={17} />
      </button>
      <button type="button" class="dock-icon dock-badge-wrap" aria-label="播放列表" title="播放列表">
        <ListMusic size={17} />
        <span class="badge">7</span>
      </button>
    </div>
  </footer>
</div>

<style>
  .player-view {
    --text: rgba(255, 255, 255, 0.96);
    --text-dim: rgba(255, 255, 255, 0.42);
    --text-muted: rgba(255, 255, 255, 0.28);
    --accent: #3ecf6e;
    --dock-bg: rgba(28, 22, 18, 0.55);

    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
    font-family:
      -apple-system,
      BlinkMacSystemFont,
      'Segoe UI',
      Roboto,
      'Helvetica Neue',
      Arial,
      sans-serif;
    color: var(--text);
    -webkit-font-smoothing: antialiased;
  }

  .audio-host {
    position: absolute;
    width: 0;
    height: 0;
    overflow: hidden;
    pointer-events: none;
  }

  .player-stage {
    flex: 1;
    min-height: 0;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 24px;
    padding: 8px 16px 16px;
    align-items: center;
  }

  @media (max-width: 900px) {
    .player-stage {
      grid-template-columns: 1fr;
      grid-template-rows: auto 1fr;
      align-items: start;
    }
  }

  /* —— 左侧视觉区 —— */
  .visual-side {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 22px;
    padding: 12px 24px;
  }

  .cover-wrap {
    position: relative;
    width: min(240px, 72vw);
    aspect-ratio: 1;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .cover-ring {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    border: 1px solid rgba(255, 255, 255, 0.22);
    box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.15);
  }

  .cover-disc {
    width: 88%;
    aspect-ratio: 1;
    border-radius: 50%;
    overflow: hidden;
    box-shadow: 0 16px 48px rgba(0, 0, 0, 0.45);
  }

  .cover-disc.spinning {
    animation: spin 18s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .cover-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .track-meta {
    text-align: center;
    max-width: 100%;
  }

  .track-title {
    margin: 0;
    font-size: clamp(1.35rem, 2.5vw, 1.75rem);
    font-weight: 700;
    letter-spacing: -0.02em;
    line-height: 1.2;
    color: var(--text);
    text-shadow: 0 1px 12px rgba(0, 0, 0, 0.35);
  }

  .track-sub {
    margin: 8px 0 0;
    font-size: 0.875rem;
    font-weight: 400;
    color: var(--text-dim);
  }

  .mini-progress {
    display: flex;
    align-items: center;
    gap: 12px;
    width: 100%;
    max-width: 300px;
  }

  .mini-time {
    font-size: 0.75rem;
    font-weight: 500;
    font-variant-numeric: tabular-nums;
    color: var(--text-dim);
    min-width: 2.5rem;
  }

  .mini-time:last-child {
    text-align: right;
  }

  .mini-track {
    flex: 1;
    height: 2px;
    background: rgba(255, 255, 255, 0.2);
    border-radius: 999px;
    cursor: pointer;
    padding: 8px 0;
    background-clip: content-box;
  }

  .mini-fill {
    height: 2px;
    border-radius: 999px;
    background: rgba(255, 255, 255, 0.85);
    pointer-events: none;
    transition: width 0.1s linear;
  }

  /* —— 右侧歌词 —— */
  .lyrics-side {
    position: relative;
    display: flex;
    flex-direction: column;
    min-height: 0;
    height: 100%;
    padding: 24px 16px;
  }

  .lyric-loading {
    position: absolute;
    top: 12px;
    right: 12px;
    color: var(--text-dim);
    animation: spin 1s linear infinite;
  }

  .status-hint {
    margin: 0 0 8px;
    text-align: center;
    font-size: 0.8rem;
    color: var(--text-dim);
  }

  .lyrics-empty {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    color: var(--text-muted);
    font-size: 0.875rem;
  }

  .lyrics-scroll {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    padding: 30% 12px;
    text-align: center;
    mask-image: linear-gradient(
      180deg,
      transparent 0%,
      #000 12%,
      #000 88%,
      transparent 100%
    );
    scrollbar-width: none;
  }

  .lyrics-scroll::-webkit-scrollbar {
    display: none;
  }

  .lyric-line {
    display: block;
    width: 100%;
    margin: 0;
    padding: 12px 8px;
    border: none;
    background: transparent;
    font-size: clamp(1.05rem, 2vw, 1.35rem);
    font-weight: 400;
    line-height: 1.65;
    color: var(--text-muted);
    opacity: 0.65;
    cursor: pointer;
    text-align: center;
    transition:
      opacity 0.28s ease,
      font-weight 0.28s ease,
      color 0.28s ease,
      font-size 0.28s ease;
  }

  .lyric-line:hover:not(.active) {
    opacity: 0.85;
    color: var(--text-dim);
  }

  .lyric-line.active {
    font-size: clamp(1.2rem, 2.4vw, 1.55rem);
    font-weight: 700;
    color: var(--text);
    opacity: 1;
    text-shadow: 0 0 24px rgba(255, 255, 255, 0.15);
  }

  /* —— 底部 Dock —— */
  .bottom-dock {
    flex-shrink: 0;
    display: grid;
    grid-template-columns: minmax(160px, 220px) 1fr minmax(200px, auto);
    align-items: center;
    gap: 16px;
    margin: 0 12px 10px;
    padding: 12px 18px 14px;
    background: var(--dock-bg);
    backdrop-filter: blur(40px) saturate(1.2);
    -webkit-backdrop-filter: blur(40px) saturate(1.2);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 14px;
    box-shadow: 0 -4px 32px rgba(0, 0, 0, 0.25);
  }

  @media (max-width: 900px) {
    .bottom-dock {
      grid-template-columns: 1fr;
      gap: 12px;
    }
  }

  .dock-left {
    display: flex;
    align-items: center;
    gap: 10px;
    min-width: 0;
  }

  .dock-thumb {
    width: 48px;
    height: 48px;
    border-radius: 6px;
    object-fit: cover;
    flex-shrink: 0;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  }

  .dock-meta {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .dock-title {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--text);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .dock-artist {
    font-size: 0.75rem;
    font-weight: 400;
    color: var(--text-dim);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .dock-center {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    min-width: 0;
  }

  .transport {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 14px;
  }

  .dock-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    padding: 0;
    border: none;
    border-radius: 50%;
    background: transparent;
    color: rgba(255, 255, 255, 0.75);
    cursor: pointer;
    transition: color 0.15s ease, background 0.15s ease;
  }

  .dock-icon:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.08);
  }

  .dock-play {
    width: 44px;
    height: 44px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.12);
    color: #fff;
    cursor: pointer;
    transition: background 0.15s ease, transform 0.15s ease;
  }

  .dock-play:hover {
    background: rgba(255, 255, 255, 0.2);
    transform: scale(1.04);
  }

  .main-progress {
    width: 100%;
    max-width: 520px;
    padding: 0 8px;
  }

  .main-track {
    position: relative;
    height: 4px;
    background: rgba(255, 255, 255, 0.15);
    border-radius: 999px;
    cursor: pointer;
    padding: 10px 0;
    background-clip: content-box;
  }

  .main-fill {
    position: relative;
    height: 4px;
    border-radius: 999px;
    background: var(--accent);
    pointer-events: none;
    transition: width 0.1s linear;
  }

  .main-thumb {
    position: absolute;
    right: 0;
    top: 50%;
    width: 12px;
    height: 12px;
    margin-top: -6px;
    margin-right: -6px;
    border-radius: 50%;
    background: #fff;
    box-shadow: 0 1px 6px rgba(0, 0, 0, 0.35);
  }

  .dock-right {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 4px;
    flex-wrap: wrap;
  }

  .stars {
    display: flex;
    align-items: center;
    gap: 1px;
    margin-right: 6px;
    color: rgba(255, 255, 255, 0.35);
  }

  .stars :global(.star-on) {
    color: rgba(255, 220, 100, 0.9);
    fill: rgba(255, 220, 100, 0.9);
  }

  .stars :global(.star-off) {
    color: rgba(255, 255, 255, 0.25);
  }

  .dock-badge-wrap {
    position: relative;
  }

  .badge {
    position: absolute;
    top: -2px;
    right: -2px;
    min-width: 14px;
    height: 14px;
    padding: 0 3px;
    border-radius: 999px;
    background: #e74c3c;
    color: #fff;
    font-size: 9px;
    font-weight: 700;
    line-height: 14px;
    text-align: center;
  }
</style>
