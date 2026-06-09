<script lang="ts">
  import { createVirtualizer } from '@tanstack/svelte-virtual';
  import { get } from 'svelte/store';
  import TrackListRow from '@/components/track/TrackListRow.svelte';
  import type { TrackItem } from '@/lib/playback/track';
  import type { PlayerTrack } from '@/stores/playback/player.svelte';

  interface Props {
    tracks: TrackItem[];
    activeId?: string | number | null;
    indexOffset?: number;
    brokenCovers?: Record<string, true>;
    coverByPath?: Record<string, string>;
    onSelect?: (track: TrackItem) => void;
    onOpenDetail?: (track: TrackItem) => void;
    onCoverError?: (url: string) => void;
    resolvePlayerTrack?: (track: TrackItem) => PlayerTrack;
    ariaLabel?: string;
    selectionMode?: boolean;
    selectedIds?: Record<string, true>;
    onToggleSelect?: (track: TrackItem, selected: boolean) => void;
    /** Show file size column (local library) */
    showSize?: boolean;
    /** 分批渲染行，避免大列表首次挂载阻塞路由切换 */
    incremental?: boolean;
    initialBatch?: number;
    batchSize?: number;
    /** 真虚拟滚动，适合 500+ 行列表 */
    virtual?: boolean;
    /** tracks 数量达到此值时自动启用 virtual（0 = 关闭）；显式 virtual={true} 仍优先 */
    virtualAutoThreshold?: number;
    virtualRowHeight?: number;
    /** When set, only resets incremental batching when this key changes (not on every tracks ref change). */
    resetKey?: string | number;
    /** Stable id for restoring incremental render limit across track array swaps (e.g. local folder tabs). */
    listId?: string;
    /** Read covers from localLibrary per row (avoids passing the whole map as a prop). */
    localCovers?: boolean;
    /** Pause infinite scroll while panel is hidden */
    paused?: boolean;
    /** When false, rows do not subscribe to per-path cover updates */
    loadCovers?: boolean;
  }

  let {
    tracks,
    activeId = null,
    indexOffset = 0,
    brokenCovers = {},
    coverByPath,
    onSelect,
    onOpenDetail,
    onCoverError,
    resolvePlayerTrack,
    ariaLabel = '歌曲列表',
    selectionMode = false,
    selectedIds = {},
    onToggleSelect,
    showSize = false,
    incremental = false,
    initialBatch = 80,
    batchSize = 80,
    virtual = false,
    virtualAutoThreshold = 500,
    virtualRowHeight = 58,
    resetKey,
    listId,
    localCovers = false,
    paused = false,
    loadCovers = true,
  }: Props = $props();

  const renderLimitsByList = new Map<string, number>();

  let renderLimit = $state(80);
  let loadMoreEl = $state<HTMLDivElement | null>(null);
  let scrollEl = $state<HTMLDivElement | null>(null);

  const useVirtual = $derived(
    tracks.length > 0 &&
      (virtual || (virtualAutoThreshold > 0 && tracks.length >= virtualAutoThreshold)),
  );
  const useIncremental = $derived(!useVirtual && incremental);
  const showPlaylistAction = $derived(Boolean(resolvePlayerTrack));
  const visibleTracks = $derived(useIncremental ? tracks.slice(0, renderLimit) : tracks);
  const hasMoreTracks = $derived(useIncremental && renderLimit < tracks.length);

  const virtualizer = createVirtualizer({
    get count() {
      return tracks.length;
    },
    getScrollElement: () => scrollEl,
    estimateSize: () => virtualRowHeight,
    overscan: 12,
  });

  // 列表在 display:none 容器里预挂载时，虚拟滚动视口高度为 0，切回页面后需重新测量
  $effect(() => {
    if (!useVirtual || !scrollEl) {
      return;
    }
    const el = scrollEl;
    tracks.length;

    const remeasure = () => {
      get(virtualizer).measure();
    };

    remeasure();
    const observer = new ResizeObserver(remeasure);
    observer.observe(el);
    return () => observer.disconnect();
  });

  $effect(() => {
    if (!useIncremental) {
      return;
    }
    if (resetKey !== undefined) {
      resetKey;
      renderLimit = initialBatch;
      if (listId) {
        renderLimitsByList.set(listId, initialBatch);
      }
      return;
    }
    if (listId !== undefined) {
      listId;
      renderLimit = renderLimitsByList.get(listId) ?? initialBatch;
      return;
    }
    tracks;
    renderLimit = initialBatch;
  });

  $effect(() => {
    if (!listId) {
      return;
    }
    renderLimitsByList.set(listId, renderLimit);
  });

  $effect(() => {
    if (!useIncremental) {
      return;
    }
    renderLimit;
    tracks.length;
    paused;

    if (paused || !loadMoreEl || !hasMoreTracks) {
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          renderLimit = Math.min(tracks.length, renderLimit + batchSize);
        }
      },
      { rootMargin: '240px 0px' },
    );

    observer.observe(loadMoreEl);
    return () => observer.disconnect();
  });

  function handleOpenDetail(e: MouseEvent, track: TrackItem) {
    e.stopPropagation();
    onOpenDetail?.(track);
  }

</script>

<div class="track-list" class:virtual-mode={useVirtual} role="table" aria-label={ariaLabel}>
  <div
    class="track-list-header"
    class:selection-mode={selectionMode}
    class:has-actions={showPlaylistAction}
    class:has-size={showSize}
    role="row"
  >
    {#if selectionMode}
      <span class="col-select" role="columnheader"></span>
    {/if}
    <span class="col-index" role="columnheader">序号</span>
    <span class="col-track" role="columnheader">歌曲</span>
    <span class="col-album" role="columnheader">专辑</span>
    <span class="col-duration" role="columnheader">时长</span>
    {#if showSize}
      <span class="col-size" role="columnheader">大小</span>
    {/if}
    {#if showPlaylistAction}
      <span class="col-actions" role="columnheader" aria-hidden="true"></span>
    {/if}
  </div>

  {#if useVirtual}
    <div class="track-list-scroll" bind:this={scrollEl}>
      <div class="track-list-virtual-spacer" style={`height: ${$virtualizer.getTotalSize()}px`}>
        {#each $virtualizer.getVirtualItems() as virtualRow (virtualRow.key)}
          {@const track = tracks[virtualRow.index]}
          <div
            class="track-row-host track-row-virtual"
            style={`height: ${virtualRow.size}px; transform: translateY(${virtualRow.start}px)`}
          >
            <TrackListRow
              {track}
              index={virtualRow.index}
              {indexOffset}
              {activeId}
              {brokenCovers}
              {localCovers}
              {loadCovers}
              {coverByPath}
              {onSelect}
              {onOpenDetail}
              {onCoverError}
              {resolvePlayerTrack}
              {selectionMode}
              {selectedIds}
              {onToggleSelect}
              {showSize}
              {showPlaylistAction}
            />
          </div>
        {/each}
      </div>
    </div>
  {:else}
  {#each visibleTracks as track, index (track.listKey ?? track.id)}
    <div class="track-row-host">
      <TrackListRow
        {track}
        {index}
        {indexOffset}
        {activeId}
        {brokenCovers}
        {localCovers}
        {loadCovers}
        {coverByPath}
        {onSelect}
        {onOpenDetail}
        {onCoverError}
        {resolvePlayerTrack}
        {selectionMode}
        {selectedIds}
        {onToggleSelect}
        {showSize}
        {showPlaylistAction}
      />
    </div>
  {/each}

  {#if hasMoreTracks}
    <div class="track-list-load-more" bind:this={loadMoreEl} aria-hidden="true"></div>
  {/if}
  {/if}
</div>

<style>
  .track-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .track-list-load-more {
    height: 1px;
    pointer-events: none;
  }

  .track-row-host {
    display: contents;
  }

  .track-list-header,
  .track-list :global(.track-row) {
    display: grid;
    grid-template-columns: 3rem minmax(0, 1fr) minmax(8rem, 28%) 4.5rem;
    align-items: center;
    column-gap: 1rem;
    padding: 0.75rem 1rem;
    text-align: left;
  }

  .track-list-header.has-actions,
  .track-list :global(.track-row.has-actions) {
    grid-template-columns: 3rem minmax(0, 1fr) minmax(8rem, 28%) 4.5rem 2.25rem;
  }

  .track-list-header.selection-mode,
  .track-list :global(.track-row.selection-mode) {
    grid-template-columns: 2rem 3rem minmax(0, 1fr) minmax(8rem, 28%) 4.5rem;
  }

  .track-list-header.selection-mode.has-actions,
  .track-list :global(.track-row.selection-mode.has-actions) {
    grid-template-columns: 2rem 3rem minmax(0, 1fr) minmax(8rem, 28%) 4.5rem 2.25rem;
  }

  .track-list-header.has-size,
  .track-list :global(.track-row.has-size) {
    grid-template-columns: 3rem minmax(0, 1fr) minmax(8rem, 24%) 4.5rem 4.5rem;
  }

  .track-list-header.has-size.has-actions,
  .track-list :global(.track-row.has-size.has-actions) {
    grid-template-columns: 3rem minmax(0, 1fr) minmax(8rem, 24%) 4.5rem 4.5rem 2.25rem;
  }

  .track-list-header.selection-mode.has-size,
  .track-list :global(.track-row.selection-mode.has-size) {
    grid-template-columns: 2rem 3rem minmax(0, 1fr) minmax(8rem, 24%) 4.5rem 4.5rem;
  }

  .track-list-header.selection-mode.has-size.has-actions,
  .track-list :global(.track-row.selection-mode.has-size.has-actions) {
    grid-template-columns: 2rem 3rem minmax(0, 1fr) minmax(8rem, 24%) 4.5rem 4.5rem 2.25rem;
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

  .track-list :global(.track-row) {
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

  .track-list :global(.track-row:focus-visible) {
    box-shadow: inset 0 0 0 2px rgba(102, 126, 234, 0.45);
  }

  .track-list :global(.track-row:hover:not(.active)) {
    background: #f5f5f5;
  }

  .track-list :global(.track-row.active) {
    background: #f0f4ff;
  }

  .track-list :global(.col-index) {
    color: #999;
    font-size: 0.875rem;
    font-variant-numeric: tabular-nums;
    text-align: center;
  }

  .track-list :global(.col-track) {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    min-width: 0;
  }

  .track-list :global(.track-meta) {
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
    min-width: 0;
  }

  .track-list :global(.track-title) {
    font-size: 0.9375rem;
    font-weight: 600;
    color: #1a1a1a;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .track-list :global(.track-title-link) {
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

  .track-list :global(.track-row.active .track-title-link) {
    color: #5b6ee8;
  }

  .track-list :global(.track-artist) {
    font-size: 0.8125rem;
    color: #999;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .track-list :global(.col-album) {
    font-size: 0.8125rem;
    color: #999;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .track-list :global(.col-duration) {
    font-size: 0.8125rem;
    color: #999;
    text-align: right;
    font-variant-numeric: tabular-nums;
  }

  .track-list :global(.col-size) {
    font-size: 0.8125rem;
    color: #999;
    text-align: right;
    white-space: nowrap;
  }

  .track-list :global(.col-actions) {
    display: flex;
    align-items: center;
    justify-content: center;
    opacity: 0;
    transition: opacity 0.15s ease;
  }

  .track-list :global(.track-row:hover .col-actions),
  .track-list :global(.track-row:focus-within .col-actions),
  .track-list :global(.track-row.active .col-actions),
  .track-list :global(.track-row:has(.track-row-action-btn[data-state='open']) .col-actions) {
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

  .track-list :global(.track-row.active .track-row-action-btn) {
    color: #9aa8f0;
  }

  .track-list :global(.track-row.active .track-row-action-btn:hover),
  .track-list :global(.track-row.active .track-row-action-btn[data-state='open']) {
    color: #5b6ee8;
    background: rgba(91, 110, 232, 0.14);
  }

  .track-list :global(.track-row.active .col-index),
  .track-list :global(.track-row.active .track-title),
  .track-list :global(.track-row.active .track-artist),
  .track-list :global(.track-row.active .col-album),
  .track-list :global(.track-row.active .col-duration) {
    color: #5b6ee8;
  }

  .track-list :global(.song-cover) {
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

  .track-list :global(.cover-img) {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  .track-list :global(.cover-overlay) {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(91, 110, 232, 0.55);
    color: #fff;
  }

  .track-list :global(.track-row.active .song-cover) {
    color: #5b6ee8;
    background: rgba(91, 110, 232, 0.12);
  }

  .track-list :global(.playing-indicator) {
    display: inline-flex;
    align-items: flex-end;
    justify-content: center;
    gap: 2px;
    height: 14px;
    width: 100%;
  }

  .track-list :global(.playing-indicator span) {
    width: 3px;
    background: #5b6ee8;
    border-radius: 1px;
    animation: bounce 0.8s ease-in-out infinite;
  }

  .track-list :global(.playing-indicator span:nth-child(1)) {
    height: 60%;
    animation-delay: 0s;
  }

  .track-list :global(.playing-indicator span:nth-child(2)) {
    height: 100%;
    animation-delay: 0.2s;
  }

  .track-list :global(.playing-indicator span:nth-child(3)) {
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

  .track-list.virtual-mode .track-list-scroll {
    max-height: min(70vh, calc(100vh - 240px));
    overflow-y: auto;
    overscroll-behavior: contain;
  }

  .track-list-virtual-spacer {
    position: relative;
    width: 100%;
  }

  .track-row-virtual {
    display: block;
    width: 100%;
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
  }
</style>
