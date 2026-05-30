<!--
  参考稿风格：左视觉右歌词；播放控制由底部全局 PlayerBar 负责。
-->
<script lang="ts">
  import { Music, LoaderCircle } from '@lucide/svelte';
  import SpectrumAnalyzer from '@/components/player/SpectrumAnalyzer.svelte';
  import { player } from '@/stores/playback/player.svelte';
  import { parseLrc, findActiveLineIndex, type LrcLine } from '@/lib/lrc';
  import { lrcRaw, lyricLoading, lyricError } from '@/stores/playback/lyrics';
  import { audioCurrentTime, audioLoading, audioError, mountAudioTo, seekAudio } from '@/stores/playback/audioEngine';
  import defaultCover from '@/assets/images/default.jpg';

  let track = $derived(player.currentSong);
  let coverSrc = $derived(track.coverUrl?.trim() || defaultCover);
  let currentTime = $derived($audioCurrentTime);

  let audioHostEl = $state<HTMLDivElement | null>(null);
  let lyricsContainerEl = $state<HTMLDivElement | null>(null);
  let lineElements = $state<Record<number, HTMLButtonElement | undefined>>({});

  let lyricLines = $derived(parseLrc($lrcRaw));
  let activeIndex = $derived(findActiveLineIndex(lyricLines, currentTime));
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

  $effect(() => {
    const host = audioHostEl;
    if (!host) return;
    return mountAudioTo(host);
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
        <div class="cover-disc" class:spinning={player.isPlaying}>
          <img src={coverSrc} alt={`${track.title} 封面`} class="cover-img" />
        </div>
      </div>

      <div class="track-meta">
        <h2 id="track-heading" class="track-title">{track.title}</h2>
        <p class="track-sub">{artistLine}</p>
      </div>

      <SpectrumAnalyzer active={player.isPlaying} tone="light" bars={44} />
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
</div>

<style>
  .player-view {
    --text: rgba(255, 255, 255, 0.96);
    --text-dim: rgba(255, 255, 255, 0.42);
    --text-muted: rgba(255, 255, 255, 0.28);
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

</style>
