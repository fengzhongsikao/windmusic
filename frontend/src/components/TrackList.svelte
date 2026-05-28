<script lang="ts">
  import { Music, Play } from '@lucide/svelte';
  import type { TrackItem } from '@/lib/track';

  interface Props {
    tracks: TrackItem[];
    activeId?: string | number | null;
    isPlaying?: boolean;
    indexOffset?: number;
    brokenCovers?: Record<string, true>;
    onSelect?: (track: TrackItem) => void;
    onCoverError?: (url: string) => void;
    ariaLabel?: string;
  }

  let {
    tracks,
    activeId = null,
    isPlaying = false,
    indexOffset = 0,
    brokenCovers = {},
    onSelect,
    onCoverError,
    ariaLabel = '歌曲列表',
  }: Props = $props();

  function isActive(track: TrackItem): boolean {
    return activeId != null && activeId === track.id;
  }

  function getCoverUrl(track: TrackItem): string {
    return track.coverUrl?.trim() ?? '';
  }

  function hasCover(track: TrackItem): boolean {
    const url = getCoverUrl(track);
    return url !== '' && !brokenCovers[url];
  }
</script>

<div class="track-list" role="table" aria-label={ariaLabel}>
  <div class="track-list-header" role="row">
    <span class="col-index" role="columnheader">序号</span>
    <span class="col-track" role="columnheader">歌曲</span>
    <span class="col-album" role="columnheader">专辑</span>
    <span class="col-duration" role="columnheader">时长</span>
  </div>

  {#each tracks as track, index (track.id)}
    <button
      type="button"
      class="track-row"
      class:active={isActive(track)}
      role="row"
      onclick={() => onSelect?.(track)}
    >
      <span class="col-index" role="cell">
        {#if isActive(track) && isPlaying}
          <span class="playing-indicator" aria-hidden="true">
            <span></span><span></span><span></span>
          </span>
        {:else}
          {indexOffset + index + 1}
        {/if}
      </span>

      <span class="col-track" role="cell">
        <span class="song-cover" aria-hidden="true">
          {#if hasCover(track)}
            <img
              src={getCoverUrl(track)}
              alt=""
              class="cover-img"
              loading="lazy"
              onerror={() => onCoverError?.(getCoverUrl(track))}
            />
            {#if isActive(track) && isPlaying}
              <span class="cover-overlay">
                <Play size={14} />
              </span>
            {/if}
          {:else if isActive(track) && isPlaying}
            <Play size={14} />
          {:else}
            <Music size={14} />
          {/if}
        </span>
        <span class="track-meta">
          <span class="track-title">{track.title}</span>
          <span class="track-artist">{track.artist}</span>
        </span>
      </span>

      <span class="col-album" role="cell">{track.album}</span>
      <span class="col-duration" role="cell">{track.duration}</span>
    </button>
  {/each}
</div>

<style>
  .track-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .track-list-header,
  .track-row {
    display: grid;
    grid-template-columns: 3rem minmax(0, 1fr) minmax(8rem, 28%) 4.5rem;
    align-items: center;
    column-gap: 1rem;
    padding: 0.75rem 1rem;
    text-align: left;
  }

  .track-list-header {
    padding-top: 0;
    padding-bottom: 0.5rem;
    font-size: 0.75rem;
    font-weight: 500;
    color: #999;
    letter-spacing: 0.02em;
  }

  .track-row {
    width: 100%;
    border: none;
    background: transparent;
    border-radius: 8px;
    cursor: pointer;
    font: inherit;
    color: inherit;
    transition: background-color 0.15s ease;
  }

  .track-row:hover:not(.active) {
    background: #f5f5f5;
  }

  .track-row.active {
    background: #f0f4ff;
  }

  .col-index {
    color: #999;
    font-size: 0.875rem;
    font-variant-numeric: tabular-nums;
    text-align: center;
  }

  .col-track {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    min-width: 0;
  }

  .track-meta {
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
    min-width: 0;
  }

  .track-title {
    font-size: 0.9375rem;
    font-weight: 600;
    color: #1a1a1a;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .track-artist {
    font-size: 0.8125rem;
    color: #999;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .col-album {
    font-size: 0.8125rem;
    color: #999;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .col-duration {
    font-size: 0.8125rem;
    color: #999;
    text-align: right;
    font-variant-numeric: tabular-nums;
  }

  .track-row.active .col-index,
  .track-row.active .track-title,
  .track-row.active .track-artist,
  .track-row.active .col-album,
  .track-row.active .col-duration {
    color: #5b6ee8;
  }

  .song-cover {
    position: relative;
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #f0f0f0;
    border-radius: 6px;
    flex-shrink: 0;
    color: #bbb;
    overflow: hidden;
  }

  .cover-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  .cover-overlay {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(91, 110, 232, 0.55);
    color: #fff;
  }

  .track-row.active .song-cover {
    color: #5b6ee8;
    background: rgba(91, 110, 232, 0.12);
  }

  .playing-indicator {
    display: inline-flex;
    align-items: flex-end;
    justify-content: center;
    gap: 2px;
    height: 14px;
    width: 100%;
  }

  .playing-indicator span {
    width: 3px;
    background: #5b6ee8;
    border-radius: 1px;
    animation: bounce 0.8s ease-in-out infinite;
  }

  .playing-indicator span:nth-child(1) {
    height: 60%;
    animation-delay: 0s;
  }

  .playing-indicator span:nth-child(2) {
    height: 100%;
    animation-delay: 0.2s;
  }

  .playing-indicator span:nth-child(3) {
    height: 40%;
    animation-delay: 0.4s;
  }

  @keyframes bounce {
    0%,
    100% {
      transform: scaleY(1);
    }
    50% {
      transform: scaleY(0.4);
    }
  }
</style>
