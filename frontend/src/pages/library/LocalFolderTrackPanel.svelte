<script lang="ts">
  import TrackList from '@/components/track/TrackList.svelte';
  import type { TrackItem } from '@/lib/track';
  import type { LocalSortOption } from '@/lib/localTrackSort';
  import { LOCAL_ALL_TAB_ID } from '@/stores/library/localLibrary.svelte';
  import { player } from '@/stores/playback/player.svelte';

  interface Props {
    tabId: string;
    tabLabel: string;
    tracks: TrackItem[];
    sortKey: LocalSortOption;
    onSelect: (track: TrackItem) => void;
  }

  let { tabId, tabLabel, tracks, sortKey, onSelect }: Props = $props();

  const displayActiveId = $derived(player.currentSong.id);
  const listId = $derived(`${tabId}:${sortKey}`);
</script>

<div class="folder-track-panel">
  {#if tracks.length === 0}
    <p class="empty-hint">
      {#if tabId === LOCAL_ALL_TAB_ID}
        暂无本地歌曲，请先添加音乐文件夹，或点击「刷新」扫描。
      {:else}
        该文件夹下暂无歌曲。
      {/if}
    </p>
  {:else}
    <TrackList
      {tracks}
      listId={listId}
      incremental
      initialBatch={100}
      batchSize={100}
      activeId={displayActiveId}
      {onSelect}
      localCovers
      loadCovers
      showSize
      ariaLabel={tabId === LOCAL_ALL_TAB_ID ? '本地歌曲列表' : `本地歌曲列表：${tabLabel}`}
    />
  {/if}
</div>

<style>
  .empty-hint {
    color: #999;
    font-size: 14px;
    margin: 0;
  }
</style>
