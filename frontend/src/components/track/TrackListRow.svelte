<script lang="ts">
  import { Music, Play } from '@lucide/svelte';
  import AddToPlaylistMenu from '@/components/playlist/AddToPlaylistMenu.svelte';
  import type { TrackItem } from '@/lib/playback/track';
  import type { PlayerTrack } from '@/stores/playback/player.svelte';
  import { player } from '@/stores/playback/player.svelte';
  import { localLibrary } from '@/stores/library/localLibrary.svelte';
  import { localDefaultCover } from '@/lib/playback/playerDefaultCovers';

  interface Props {
    track: TrackItem;
    index: number;
    indexOffset?: number;
    activeId?: string | number | null;
    brokenCovers?: Record<string, true>;
    /** Read cover from localLibrary.coverByPath[filePath] with per-row reactivity */
    localCovers?: boolean;
    loadCovers?: boolean;
    coverByPath?: Record<string, string>;
    onSelect?: (track: TrackItem) => void;
    onOpenDetail?: (track: TrackItem) => void;
    onCoverError?: (url: string) => void;
    resolvePlayerTrack?: (track: TrackItem) => PlayerTrack;
    selectionMode?: boolean;
    selectedIds?: Record<string, true>;
    onToggleSelect?: (track: TrackItem, selected: boolean) => void;
    showSize?: boolean;
    showPlaylistAction?: boolean;
  }

  let {
    track,
    index,
    indexOffset = 0,
    activeId = null,
    brokenCovers = {},
    localCovers = false,
    loadCovers = true,
    coverByPath,
    onSelect,
    onOpenDetail,
    onCoverError,
    resolvePlayerTrack,
    selectionMode = false,
    selectedIds = {},
    onToggleSelect,
    showSize = false,
    showPlaylistAction = false,
  }: Props = $props();

  const rowKey = $derived(String(track.listKey ?? track.id));
  const filePath = $derived(rowKey);

  const coverUrl = $derived.by(() => {
    const direct = track.coverUrl?.trim();
    if (direct) {
      return direct;
    }
    if (!loadCovers) {
      return '';
    }

    const lookupKeys = [filePath, String(track.id).trim()].filter(
      (key, index, keys) => key && keys.indexOf(key) === index,
    );

    if (coverByPath) {
      for (const key of lookupKeys) {
        const fromMap = coverByPath[key]?.trim();
        if (fromMap) {
          return fromMap;
        }
      }
    }
    if (localCovers) {
      for (const key of lookupKeys) {
        void localLibrary.coverTickByPath[key];
        const fromStore = localLibrary.coverByPath[key]?.trim();
        if (fromStore) {
          return fromStore;
        }
      }
      return localDefaultCover;
    }
    return '';
  });

  const hasCover = $derived(coverUrl !== '' && !brokenCovers[coverUrl]);
  const active = $derived(activeId != null && activeId === track.id);
  const showPlaying = $derived(active ? player.isPlaying : false);
  const sizeLabel = $derived(track.size?.trim() || '—');
</script>

<div
  class="track-row"
  class:active
  class:selection-mode={selectionMode}
  class:has-actions={showPlaylistAction}
  class:has-size={showSize}
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
    {#if active && showPlaying}
      <span class="playing-indicator" aria-hidden="true">
        <span></span><span></span><span></span>
      </span>
    {:else}
      {indexOffset + index + 1}
    {/if}
  </span>

  <span class="col-track" role="cell">
    <span class="song-cover" aria-hidden="true">
      {#if hasCover}
        <img
          src={coverUrl}
          alt=""
          class="cover-img"
          loading="lazy"
          decoding="async"
          onerror={() => onCoverError?.(coverUrl)}
        />
        {#if active && showPlaying}
          <span class="cover-overlay">
            <Play size={14} />
          </span>
        {/if}
      {:else if active && showPlaying}
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
          onclick={(e) => {
            e.stopPropagation();
            onOpenDetail(track);
          }}
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
  {#if showSize}
    <span class="col-size" role="cell">{sizeLabel}</span>
  {/if}
  {#if resolvePlayerTrack}
    <span class="col-actions" role="cell">
      <AddToPlaylistMenu
        track={resolvePlayerTrack(track)}
        triggerClass="track-row-action-btn"
        placement="bottom-end"
      />
    </span>
  {:else if showPlaylistAction}
    <span class="col-actions" role="cell" aria-hidden="true"></span>
  {/if}
</div>
