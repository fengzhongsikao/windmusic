<script lang="ts">
  import TrackList from '@/components/TrackList.svelte';
  import type { TrackItem } from '@/lib/track';
  import { LOCAL_ALL_TAB_ID, localLibrary } from '@/stores/localLibrary.svelte';
  import { player } from '@/stores/player.svelte';

  interface Props {
    tabId: string;
    tabLabel: string;
    visible: boolean;
    onSelect: (track: TrackItem) => void;
  }

  let { tabId, tabLabel, visible, onSelect }: Props = $props();

  const tracks = $derived(localLibrary.tracksByTab[tabId] ?? []);

  let displayActiveId = $state<string | number | null>(null);
  let displayPlaying = $state(false);

  $effect(() => {
    if (!visible) {
      return;
    }
    displayActiveId = player.currentSong.id;
    displayPlaying = player.isPlaying;
  });
</script>

<div class="folder-track-panel" class:is-visible={visible}>
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
      listId={tabId}
      activeId={displayActiveId}
      isPlaying={displayPlaying}
      {onSelect}
      localCovers
      showSize
      ariaLabel={tabId === LOCAL_ALL_TAB_ID ? '本地歌曲列表' : `本地歌曲列表：${tabLabel}`}
    />
  {/if}
</div>

<style>
  .folder-track-panel {
    display: none;
  }

  .folder-track-panel.is-visible {
    display: block;
  }

  .folder-track-panel:not(.is-visible) {
    content-visibility: hidden;
    contain: strict;
    pointer-events: none;
  }

  .empty-hint {
    color: #999;
    font-size: 14px;
    margin: 0;
  }
</style>
