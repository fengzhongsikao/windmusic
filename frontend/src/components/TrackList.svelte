<script lang="ts">
  import { Music, Play } from '@lucide/svelte';
  import AddToPlaylistMenu from '@/components/AddToPlaylistMenu.svelte';
  import type { TrackItem } from '@/lib/track';
  import type { PlayerTrack } from '@/stores/player.svelte';

  interface Props {
    tracks: TrackItem[];
    activeId?: string | number | null;
    isPlaying?: boolean;
    indexOffset?: number;
    brokenCovers?: Record<string, true>;
    onSelect?: (track: TrackItem) => void;
    onOpenDetail?: (track: TrackItem) => void;
    onCoverError?: (url: string) => void;
    resolvePlayerTrack?: (track: TrackItem) => PlayerTrack;
    ariaLabel?: string;
    selectionMode?: boolean;
    selectedIds?: Record<string, true>;
    onToggleSelect?: (track: TrackItem, selected: boolean) => void;
  }

  let {
    tracks,
    activeId = null,
    isPlaying = false,
    indexOffset = 0,
    brokenCovers = {},
    onSelect,
    onOpenDetail,
    onCoverError,
    resolvePlayerTrack,
    ariaLabel = '歌曲列表',
    selectionMode = false,
    selectedIds = {},
    onToggleSelect,
  }: Props = $props();

  const showPlaylistAction = $derived(Boolean(resolvePlayerTrack));

  function handleOpenDetail(e: MouseEvent, track: TrackItem) {
    e.stopPropagation();
    onOpenDetail?.(track);
  }

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
  <div
    class="track-list-header"
    class:selection-mode={selectionMode}
    class:has-actions={showPlaylistAction}
    role="row"
  >
    {#if selectionMode}
      <span class="col-select" role="columnheader"></span>
    {/if}
    <span class="col-index" role="columnheader">序号</span>
    <span class="col-track" role="columnheader">歌曲</span>
    <span class="col-album" role="columnheader">专辑</span>
    <span class="col-duration" role="columnheader">时长</span>
    {#if showPlaylistAction}
      <span class="col-actions" role="columnheader" aria-hidden="true"></span>
    {/if}
  </div>

  {#each tracks as track, index (track.listKey ?? track.id)}
    <div
      class="track-row"
      class:active={isActive(track)}
      class:selection-mode={selectionMode}
      class:has-actions={showPlaylistAction}
      role="row"
      tabindex="0"
      onclick={() => {
        if (!selectionMode) {
          onSelect?.(track);
        }
      }}
      onkeydown={(e) => {
        if (selectionMode) return;
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          onSelect?.(track);
        }
      }}
    >
      {#if selectionMode}
        <span class="col-select" role="cell">
          <input
            type="checkbox"
            checked={Boolean(selectedIds[track.listKey ?? String(track.id)])}
            aria-label={`选择 ${track.title}`}
            onclick={(e) => e.stopPropagation()}
            oninput={(e) => onToggleSelect?.(track, (e.currentTarget as HTMLInputElement).checked)}
          />
        </span>
      {/if}
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
          {#if onOpenDetail}
            <button
              type="button"
              class="track-title track-title-link"
              onclick={(e) => handleOpenDetail(e, track)}
              title="查看歌曲详情"
            >
              {track.title}
            </button>
          {:else}
            <span class="track-title">{track.title}</span>
          {/if}
          <span class="track-artist">{track.artist}</span>
        </span>
      </span>

      <span class="col-album" role="cell">{track.album}</span>
      <span class="col-duration" role="cell">{track.duration}</span>
      {#if resolvePlayerTrack}
        <span class="col-actions" role="cell">
          <AddToPlaylistMenu
            track={resolvePlayerTrack(track)}
            triggerClass="track-row-action-btn"
            placement="bottom-end"
          />
        </span>
      {/if}
    </div>
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

  .track-list-header.has-actions,
  .track-row.has-actions {
    grid-template-columns: 3rem minmax(0, 1fr) minmax(8rem, 28%) 4.5rem 2.25rem;
  }

  .track-list-header.selection-mode,
  .track-row.selection-mode {
    grid-template-columns: 2rem 3rem minmax(0, 1fr) minmax(8rem, 28%) 4.5rem;
  }

  .track-list-header.selection-mode.has-actions,
  .track-row.selection-mode.has-actions {
    grid-template-columns: 2rem 3rem minmax(0, 1fr) minmax(8rem, 28%) 4.5rem 2.25rem;
  }

  .col-select {
    display: inline-flex;
    justify-content: center;
    align-items: center;
  }

  .col-select input[type='checkbox'] {
    width: 16px;
    height: 16px;
    accent-color: #667eea;
    cursor: pointer;
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
    outline: none;
  }

  .track-row:focus-visible {
    box-shadow: inset 0 0 0 2px rgba(102, 126, 234, 0.45);
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

  .track-title-link {
    display: block;
    width: 100%;
    padding: 0;
    border: none;
    background: none;
    font: inherit;
    text-align: left;
    cursor: pointer;
  }

  .track-title-link:hover {
    color: #667eea;
    text-decoration: underline;
  }

  .track-row.active .track-title-link {
    color: #5b6ee8;
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

  .col-actions {
    display: flex;
    align-items: center;
    justify-content: center;
    opacity: 0;
    transition: opacity 0.15s ease;
  }

  .track-row:hover .col-actions,
  .track-row:focus-within .col-actions,
  .track-row.active .col-actions {
    opacity: 1;
  }

  :global(.track-row-action-btn) {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    padding: 0;
    border: none;
    border-radius: 6px;
    background: transparent;
    color: #bbb;
    cursor: pointer;
    transition:
      color 0.15s ease,
      background-color 0.15s ease;
  }

  :global(.track-row-action-btn:hover),
  :global(.track-row-action-btn[data-state='open']) {
    color: #667eea;
    background: rgba(102, 126, 234, 0.1);
  }

  .track-row.active :global(.track-row-action-btn) {
    color: #9aa8f0;
  }

  .track-row.active :global(.track-row-action-btn:hover),
  .track-row.active :global(.track-row-action-btn[data-state='open']) {
    color: #5b6ee8;
    background: rgba(91, 110, 232, 0.14);
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
