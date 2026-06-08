<!--
  参考稿风格：左视觉右歌词；播放控制由底部全局 PlayerBar 负责。
-->
<script lang="ts">
  import { Music, LoaderCircle } from '@lucide/svelte';
  import TrackWaveform from '@/components/player/TrackWaveform.svelte';
  import { player } from '@/stores/playback/player.svelte';
  import { parseLrc, findActiveLineIndex, type LrcLine } from '@/lib/playback/lrc';
  import { lrcRaw, lyricLoading, lyricError } from '@/stores/playback/lyrics';
  import { audioCurrentTime, audioLoading, audioError, seekAudio } from '@/stores/playback/audioEngine';
  import { playerUiSettings } from '@/stores/playback/playerSettings.svelte';
  import defaultCover from '@/assets/images/default.jpg';

  let track = $derived(player.currentSong);
  let coverSrc = $derived(track.coverUrl?.trim() || defaultCover);
  let currentTime = $derived($audioCurrentTime);

  let lyricsContainerEl = $state<HTMLDivElement | null>(null);
  let lineElements = $state<Record<number, HTMLButtonElement | undefined>>({});

  let lyricLines = $derived(parseLrc($lrcRaw));
  let activeIndex = $derived(findActiveLineIndex(lyricLines, currentTime));
  let artistLine = $derived(
    [track.artist, track.album].filter(Boolean).join(' - ') || track.artist,
  );
  let hideLyrics = $derived(playerUiSettings.detailHideLyrics);
  let hideVisual = $derived(playerUiSettings.detailHideVisual);
  let coverShape = $derived(playerUiSettings.detailCoverShape);
  let coverSpin = $derived(
    playerUiSettings.detailCoverSpin && coverShape === 'round',
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
  <div
    class="player-stage"
    class:visual-only={hideLyrics && !hideVisual}
    class:lyrics-only={hideVisual && !hideLyrics}
  >
    {#if !hideVisual}
    <aside class="visual-side" aria-labelledby="track-heading">
      <div class="cover-wrap" class:square={coverShape === 'square'}>
        {#if coverShape === 'round'}
          <div class="cover-ring" aria-hidden="true"></div>
        {/if}
        <div
          class="cover-art"
          class:round={coverShape === 'round'}
          class:square={coverShape === 'square'}
          class:spinning={player.isPlaying && coverSpin}
        >
          <img src={coverSrc} alt={`${track.title} 封面`} class="cover-img" />
        </div>
      </div>

      <div class="track-meta">
        <h2 id="track-heading" class="track-title">{track.title}</h2>
        <p class="track-sub">{artistLine}</p>
      </div>

      <!-- 封面下方实时波浪：TrackWaveform（AnalyserNode + Canvas，仅此刻播放片段） -->
      <TrackWaveform tone="light" />
    </aside>
    {/if}

    {#if !hideLyrics}
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
    {/if}
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

    .player-stage.visual-only,
    .player-stage.lyrics-only {
      grid-template-rows: 1fr;
      align-items: center;
    }
  }

  .player-stage.visual-only,
  .player-stage.lyrics-only {
    grid-template-columns: 1fr;
  }

  .player-stage.visual-only .visual-side {
    justify-content: center;
    height: 100%;
    padding: 24px 32px;
  }

  .player-stage.visual-only .cover-wrap {
    width: min(300px, 78vw);
  }

  .player-stage.lyrics-only .lyrics-side {
    padding: 16px 24px;
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

  .cover-wrap.square {
    width: min(260px, 76vw);
  }

  .cover-ring {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    border: 1px solid rgba(255, 255, 255, 0.22);
    box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.15);
  }

  .cover-art {
    overflow: hidden;
    box-shadow: 0 16px 48px rgba(0, 0, 0, 0.45);
  }

  .cover-art.round {
    width: 88%;
    aspect-ratio: 1;
    border-radius: 50%;
  }

  .cover-art.square {
    width: 100%;
    aspect-ratio: 1;
    border-radius: 10px;
    border: 1px solid rgba(255, 255, 255, 0.12);
  }

  .cover-art.spinning {
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
