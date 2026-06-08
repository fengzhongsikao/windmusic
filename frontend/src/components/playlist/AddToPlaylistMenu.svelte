<!--
  添加到歌单：Skeleton Menu 下拉面板（与播放队列菜单一致）。
-->
<script lang="ts">
  import { ListPlus, Music, CirclePlus } from '@lucide/svelte';
  import { Menu, Portal } from '@skeletonlabs/skeleton-svelte';
  import PlaylistCreateFields from '@/components/playlist/PlaylistCreateFields.svelte';
  import {
    addSongToPlaylist,
    playlistActionErrorMessage,
    playlistContainsSong,
  } from '@/lib/library/playlists';
  import { toFavoriteSong } from '@/lib/wails/wailsPlayer';
  import type { PlayerTrack } from '@/stores/playback/player.svelte';
  import { playlistsState } from '@/stores/library/playlistsStore.svelte';
  import { success, error as toastError } from '@/stores/ui/toast';

  interface Props {
    track: PlayerTrack;
    immersive?: boolean;
    triggerClass?: string;
    disabled?: boolean;
    placement?: 'top-end' | 'bottom-end' | 'right-start';
  }

  let {
    track,
    immersive = false,
    triggerClass = 'ctrl-btn small',
    disabled = false,
    placement = 'top-end',
  }: Props = $props();

  let addingId = $state('');
  let menuOpen = $state(false);
  let createFields = $state<PlaylistCreateFields | null>(null);

  const playlists = $derived(playlistsState.items);
  const loading = $derived(!playlistsState.loaded);
  const song = $derived(toFavoriteSong(track));
  const hasPlayableTrack = $derived(Boolean(track.title?.trim()));
  const playlistCount = $derived(playlists.length);

  function closeMenu() {
    menuOpen = false;
    createFields?.reset();
  }

  async function handleAdd(playlistId: string) {
    if (addingId || !hasPlayableTrack) return;
    addingId = playlistId;
    try {
      await addSongToPlaylist(playlistId, song);
      const pl = playlists.find((item) => item.id === playlistId);
      success(pl ? `已添加到「${pl.name}」` : '已添加到歌单');
      closeMenu();
    } catch (err) {
      toastError(playlistActionErrorMessage(err, '添加失败'));
    } finally {
      addingId = '';
    }
  }

  function handleSelect(details: { value: string }) {
    if (details.value.startsWith('add:')) {
      void handleAdd(details.value.slice(4));
    }
  }

  function handleOpenChange(details: { open: boolean }) {
    menuOpen = details.open;
    if (!details.open) {
      createFields?.reset();
    }
  }

  function handlePlaylistCreated() {
    closeMenu();
  }
</script>

<Menu
  class="add-playlist-menu-root"
  open={menuOpen}
  closeOnSelect={false}
  aria-label="添加到歌单"
  positioning={{ placement, gutter: 10 }}
  onSelect={handleSelect}
  onOpenChange={handleOpenChange}
>
  <Menu.Trigger
    class="{triggerClass} add-playlist-trigger"
    title="添加到歌单"
    aria-label="添加到歌单"
    {disabled}
    onclick={(e) => e.stopPropagation()}
  >
    <ListPlus size={16} />
  </Menu.Trigger>

  <Portal>
    <Menu.Positioner>
      <Menu.Content class="add-playlist-menu-content card {immersive ? 'immersive' : ''}">
        <Menu.ItemGroup>
          <Menu.ItemGroupLabel class="add-playlist-menu-header">
            <span class="add-playlist-menu-title">
              添加到歌单
              <span class="add-playlist-menu-count">{playlistCount} 个</span>
            </span>
          </Menu.ItemGroupLabel>

          <div class="add-playlist-menu-body">
            {#if loading && playlists.length === 0}
              <p class="add-playlist-empty">加载中…</p>
            {:else if playlists.length === 0}
              <p class="add-playlist-empty">还没有歌单，可在下方新建</p>
            {:else}
              {#each playlists as pl (pl.id)}
                {@const contained = playlistContainsSong(pl, song)}
                <Menu.Item
                  value={`add:${pl.id}`}
                  closeOnSelect={false}
                  class="add-playlist-menu-item"
                  disabled={contained || addingId === pl.id || !hasPlayableTrack}
                >
                  <span class="add-playlist-icon" aria-hidden="true">
                    <Music size={14} />
                  </span>
                  <span class="add-playlist-meta">
                    <Menu.ItemText class="add-playlist-name">{pl.name}</Menu.ItemText>
                    <span class="add-playlist-song-count">{pl.songCount} 首</span>
                  </span>
                  {#if contained}
                    <span class="add-playlist-badge">已添加</span>
                  {:else if addingId === pl.id}
                    <span class="add-playlist-badge">添加中…</span>
                  {/if}
                </Menu.Item>
              {/each}
            {/if}
          </div>
        </Menu.ItemGroup>

        <Menu.Separator class="add-playlist-separator" />

        <Menu.ItemGroup>
          <Menu.ItemGroupLabel class="add-playlist-create-header">
            <CirclePlus size={14} />
            <span>新建歌单</span>
          </Menu.ItemGroupLabel>
        </Menu.ItemGroup>
        <PlaylistCreateFields bind:this={createFields} onCreated={handlePlaylistCreated} />
      </Menu.Content>
    </Menu.Positioner>
  </Portal>
</Menu>

<style>
  :global(.add-playlist-menu-content) {
    width: min(320px, calc(100vw - 280px));
    max-height: min(420px, calc(100vh - 160px));
    padding: 0;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    z-index: 200;
  }

  :global(.add-playlist-menu-content.immersive) {
    background: rgba(28, 22, 20, 0.96);
    border-color: rgba(255, 255, 255, 0.1);
    color: rgba(255, 255, 255, 0.92);
  }

  :global(.add-playlist-menu-header) {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 12px 14px 8px;
    margin: 0;
    border-bottom: 1px solid #f0f0f0;
  }

  :global(.add-playlist-menu-content.immersive .add-playlist-menu-header) {
    border-bottom-color: rgba(255, 255, 255, 0.08);
  }

  :global(.add-playlist-menu-title) {
    display: flex;
    align-items: baseline;
    gap: 8px;
    font-size: 14px;
    font-weight: 600;
  }

  :global(.add-playlist-menu-count) {
    font-size: 12px;
    font-weight: 400;
    color: #999;
  }

  :global(.add-playlist-menu-content.immersive .add-playlist-menu-count) {
    color: rgba(255, 255, 255, 0.45);
  }

  .add-playlist-menu-body {
    overflow-y: auto;
    max-height: min(240px, calc(100vh - 280px));
    padding: 4px 6px;
  }

  :global(.add-playlist-empty) {
    margin: 0;
    padding: 16px 12px 20px;
    font-size: 13px;
    line-height: 1.6;
    color: #999;
    text-align: center;
  }

  :global(.add-playlist-menu-content.immersive .add-playlist-empty) {
    color: rgba(255, 255, 255, 0.45);
  }

  :global(.add-playlist-menu-content .add-playlist-menu-item) {
    display: grid;
    grid-template-columns: 2.25rem minmax(0, 1fr) auto;
    align-items: center;
    gap: 8px;
    padding: 8px 10px;
    cursor: pointer;
    border-radius: 8px;
  }

  :global(.add-playlist-menu-content .add-playlist-menu-item[data-disabled]) {
    opacity: 0.55;
    cursor: not-allowed;
  }

  :global(.add-playlist-icon) {
    width: 36px;
    height: 36px;
    border-radius: 6px;
    background: #f0f0f0;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #bbb;
    flex-shrink: 0;
  }

  :global(.add-playlist-menu-content.immersive .add-playlist-icon) {
    background: rgba(255, 255, 255, 0.08);
    color: rgba(255, 255, 255, 0.45);
  }

  :global(.add-playlist-meta) {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  :global(.add-playlist-menu-content .add-playlist-name) {
    font-size: 13px;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  :global(.add-playlist-song-count) {
    font-size: 11px;
    color: #999;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  :global(.add-playlist-menu-content.immersive .add-playlist-song-count) {
    color: rgba(255, 255, 255, 0.45);
  }

  :global(.add-playlist-badge) {
    font-size: 11px;
    color: #667eea;
    flex-shrink: 0;
  }

  :global(.add-playlist-separator) {
    margin: 4px 0;
    border-color: #f0f0f0;
  }

  :global(.add-playlist-menu-content.immersive .add-playlist-separator) {
    border-color: rgba(255, 255, 255, 0.08);
  }

  :global(.add-playlist-create-header) {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 14px 0;
    margin: 0;
    font-size: 12px;
    font-weight: 600;
    color: #667eea;
  }

  :global(.add-playlist-menu-trigger[data-state='open']) {
    color: #667eea;
  }
</style>
