<!--
  侧栏新建歌单：Skeleton Menu 下拉面板。
-->
<script lang="ts">
  import { CirclePlus } from '@lucide/svelte';
  import { Menu, Portal } from '@skeletonlabs/skeleton-svelte';
  import PlaylistCreateFields from '@/components/PlaylistCreateFields.svelte';

  interface Props {
    onCreated?: () => void;
  }

  let { onCreated }: Props = $props();

  let createFields = $state<PlaylistCreateFields | null>(null);

  function handleOpenChange(details: { open: boolean }) {
    if (details.open) {
      queueMicrotask(() => createFields?.focusInput());
    } else {
      createFields?.reset();
    }
  }

  function handleCreated() {
    createFields?.reset();
    onCreated?.();
  }
</script>

<Menu
  class="create-playlist-menu-root"
  closeOnSelect={false}
  aria-label="新建歌单"
  positioning={{ placement: 'bottom-start', gutter: 6 }}
  onOpenChange={handleOpenChange}
>
  <Menu.Trigger class="add-playlist-btn" title="新建歌单" aria-label="新建歌单">
    <CirclePlus size={16} />
  </Menu.Trigger>

  <Portal>
    <Menu.Positioner>
      <Menu.Content class="playlist-menu-content card">
        <Menu.ItemGroup>
          <Menu.ItemGroupLabel class="playlist-menu-header">
            <span class="playlist-menu-title">新建歌单</span>
          </Menu.ItemGroupLabel>
        </Menu.ItemGroup>
        <PlaylistCreateFields bind:this={createFields} onCreated={handleCreated} />
      </Menu.Content>
    </Menu.Positioner>
  </Portal>
</Menu>

<style>
  :global(.create-playlist-menu-root .add-playlist-btn) {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 2px;
    border: none;
    background: transparent;
    color: #999;
    cursor: pointer;
    border-radius: 4px;
    transition: all 0.2s ease;
  }

  :global(.create-playlist-menu-root .add-playlist-btn:hover),
  :global(.create-playlist-menu-root .add-playlist-btn[data-state='open']) {
    color: #667eea;
    background: rgba(102, 126, 234, 0.08);
  }

  :global(.playlist-menu-content) {
    width: min(260px, calc(100vw - 48px));
    padding: 0;
    overflow: hidden;
    z-index: 200;
  }

  :global(.playlist-menu-header) {
    padding: 12px 14px 4px;
    margin: 0;
    border-bottom: 1px solid #f0f0f0;
  }

  :global(.playlist-menu-title) {
    font-size: 14px;
    font-weight: 600;
    color: #333;
  }
</style>
