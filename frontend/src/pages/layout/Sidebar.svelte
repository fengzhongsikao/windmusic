<!--
  左侧导航：主导航、我的音乐、歌单列表；通过 hash 路由高亮当前项。
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import { Music, House, ListMusic, Trophy, Heart, Clock, FolderOpen, Trash2 } from '@lucide/svelte';
  import { link, push } from 'svelte-spa-router';
  import CreatePlaylistMenu from '@/components/playlist/CreatePlaylistMenu.svelte';
  import {
    deleteUserPlaylist,
    playlistActionErrorMessage,
    type UserPlaylist,
  } from '@/lib/library/playlists';
  import { playlistsState } from '@/stores/library/playlistsStore.svelte';
  import { error as toastError } from '@/stores/ui/toast';

  const menuItems = [
    { id: 'home', label: '首页', icon: House, path: '/' },
    // { id: 'recommend', label: '推荐歌单', icon: ListMusic, path: '/recommend' },
    // { id: 'ranking', label: '排行榜', icon: Trophy, path: '/ranking' },
  ];

  const libraryItems = [
    { id: 'favorites', label: '我喜欢的音乐', icon: Heart, path: '/favorites' },
    { id: 'recent', label: '最近播放', icon: Clock, path: '/recent' },
    { id: 'local', label: '本地音乐', icon: FolderOpen, path: '/local' },
  ];

  let deletingId = $state('');
  let pendingDelete = $state<UserPlaylist | null>(null);

  const playlists = $derived(playlistsState.items);

  let currentPath = $state(window.location.hash.slice(1) || '/');

  function syncCurrentPath() {
    currentPath = window.location.hash.slice(1) || '/';
  }

  function isActive(path: string): boolean {
    if (path === '/') {
      return currentPath === '/' || currentPath === '/discover' || currentPath === '';
    }
    return currentPath === path;
  }

  function isPlaylistActive(id: string): boolean {
    return currentPath === `/playlist/${id}`;
  }

  function playlistPath(id: string): string {
    return `/playlist/${id}`;
  }

  function handleNavigation(path: string) {
    currentPath = path;
  }

  function openDeleteDialog(pl: UserPlaylist, e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();
    if (deletingId) return;
    pendingDelete = pl;
  }

  function closeDeleteDialog() {
    if (deletingId) return;
    pendingDelete = null;
  }

  async function confirmDeletePlaylist() {
    const pl = pendingDelete;
    if (!pl || deletingId) return;

    deletingId = pl.id;
    try {
      await deleteUserPlaylist(pl.id);
      pendingDelete = null;
      if (isPlaylistActive(pl.id)) {
        push('/');
      }
    } catch (err) {
      toastError(playlistActionErrorMessage(err, '删除歌单失败'));
    } finally {
      deletingId = '';
    }
  }

  onMount(() => {
    syncCurrentPath();
    window.addEventListener('hashchange', syncCurrentPath);
    return () => {
      window.removeEventListener('hashchange', syncCurrentPath);
    };
  });
</script>

<div class="sidebar">
  <div class="logo">
    <span class="logo-icon">
      <Music size={24} />
    </span>
    <span class="logo-text">WindMusic</span>
  </div>

  <div class="nav-section">
    <div class="section-title">在线音乐</div>
    {#each menuItems as item}
      <a
        href={item.path}
        use:link
        class="nav-item"
        class:active={isActive(item.path)}
        onclick={() => handleNavigation(item.path)}
      >
        <span class="nav-icon">
          <item.icon size={18} />
        </span>
        <span class="nav-label">{item.label}</span>
      </a>
    {/each}
  </div>

  <div class="nav-section">
    <div class="section-title">我的音乐</div>
    {#each libraryItems as item}
      <a
        href={item.path}
        use:link
        class="nav-item"
        class:active={isActive(item.path)}
        onclick={() => handleNavigation(item.path)}
      >
        <span class="nav-icon">
          <item.icon size={18} />
        </span>
        <span class="nav-label">{item.label}</span>
      </a>
    {/each}
  </div>

  <div class="nav-section">
    <div class="section-header">
      <div class="section-title">创建的歌单</div>
      <CreatePlaylistMenu />
    </div>
    {#if playlists.length === 0}
      <p class="playlist-empty">点击右侧 + 创建歌单</p>
    {:else}
      {#each playlists as pl (pl.id)}
        <div
          class="playlist-row"
          class:active={isPlaylistActive(pl.id)}
          class:deleting={deletingId === pl.id}
        >
          <a
            href={playlistPath(pl.id)}
            use:link
            class="nav-item playlist-item"
            class:active={isPlaylistActive(pl.id)}
            title={pl.name}
            onclick={() => handleNavigation(playlistPath(pl.id))}
          >
            <span class="nav-icon">
              <Music size={16} />
            </span>
            <span class="nav-label">{pl.name}</span>
            <span class="playlist-count">{pl.songCount}</span>
          </a>
          <button
            type="button"
            class="playlist-delete-btn"
            title="删除歌单"
            aria-label={`删除歌单 ${pl.name}`}
            disabled={deletingId === pl.id}
            onclick={(e) => openDeleteDialog(pl, e)}
            onmousedown={(e) => e.stopPropagation()}
          >
            <Trash2 size={14} />
          </button>
        </div>
      {/each}
    {/if}
  </div>
</div>

{#if pendingDelete}
  <div class="dialog-backdrop" role="presentation" onclick={closeDeleteDialog}>
    <div
      class="alert-dialog"
      role="alertdialog"
      aria-modal="true"
      aria-labelledby="delete-playlist-sidebar-title"
      onclick={(e) => e.stopPropagation()}
    >
      <h3 id="delete-playlist-sidebar-title">删除歌单？</h3>
      <p>将永久删除「{pendingDelete.name}」及其中的全部歌曲，此操作不可恢复。</p>
      <div class="dialog-actions">
        <button type="button" class="btn dialog-btn" onclick={closeDeleteDialog} disabled={Boolean(deletingId)}>
          取消
        </button>
        <button
          type="button"
          class="btn dialog-btn danger"
          disabled={Boolean(deletingId)}
          onclick={() => void confirmDeletePlaylist()}
        >
          {deletingId ? '删除中…' : '确定删除'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .sidebar {
    width: 220px;
    height: 100%;
    background: #fafafa;
    color: #333;
    display: flex;
    flex-direction: column;
    overflow-y: auto;
    user-select: none;
    border-right: 1px solid #eee;
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 20px 20px 16px;
    font-size: 18px;
    font-weight: 700;
  }

  .logo-icon {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .logo-text {
    background: linear-gradient(135deg, #667eea, #764ba2);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .nav-section {
    padding: 8px 12px;
  }

  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px 4px;
  }

  .section-title {
    font-size: 11px;
    text-transform: uppercase;
    color: #999;
    padding: 8px 12px 4px;
    letter-spacing: 1px;
  }

  .section-header .section-title {
    padding: 0;
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 12px;
    width: 100%;
    padding: 10px 12px;
    border: none;
    background: transparent;
    color: #666;
    font-size: 14px;
    cursor: pointer;
    border-radius: 8px;
    transition: all 0.2s ease;
    text-align: left;
    text-decoration: none;
  }

  .nav-item:hover {
    background: rgba(0, 0, 0, 0.04);
    color: #333;
  }

  .nav-item.active {
    background: rgba(102, 126, 234, 0.1);
    color: #667eea;
  }

  .nav-icon {
    width: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .nav-label {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .playlist-row {
    display: flex;
    align-items: center;
    gap: 2px;
    border-radius: 8px;
  }

  .playlist-row:hover {
    background: rgba(0, 0, 0, 0.04);
  }

  .playlist-row.active {
    background: rgba(102, 126, 234, 0.1);
  }

  .playlist-row.deleting {
    opacity: 0.6;
    pointer-events: none;
  }

  .playlist-item {
    flex: 1;
    min-width: 0;
    font-size: 13px;
    padding: 8px 4px 8px 12px;
  }

  .playlist-row:hover .playlist-item,
  .playlist-row.active .playlist-item {
    background: transparent;
  }

  .playlist-count {
    margin-left: auto;
    font-size: 11px;
    color: #aaa;
    flex-shrink: 0;
  }

  .playlist-row.active .playlist-count {
    color: #667eea;
  }

  .playlist-delete-btn {
    position: relative;
    z-index: 2;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    margin-right: 4px;
    padding: 0;
    border: none;
    border-radius: 6px;
    background: transparent;
    color: #bbb;
    cursor: pointer;
    flex-shrink: 0;
    opacity: 0.45;
    pointer-events: auto;
    transition:
      opacity 0.15s ease,
      color 0.15s ease,
      background 0.15s ease;
  }

  .playlist-row:hover .playlist-delete-btn,
  .playlist-delete-btn:focus-visible {
    opacity: 1;
  }

  .playlist-delete-btn:hover:not(:disabled) {
    color: #e74c3c;
    background: rgba(231, 76, 60, 0.08);
  }

  .playlist-delete-btn:disabled {
    cursor: not-allowed;
  }

  .playlist-empty {
    margin: 0;
    padding: 4px 12px 8px;
    font-size: 12px;
    color: #aaa;
    line-height: 1.4;
  }

  .dialog-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.35);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 300;
    padding: 16px;
  }

  .alert-dialog {
    width: min(380px, 100%);
    border-radius: 14px;
    background: #fff;
    padding: 18px;
    box-shadow: 0 12px 30px rgba(0, 0, 0, 0.16);
  }

  .alert-dialog h3 {
    margin: 0;
    font-size: 18px;
    color: #111827;
  }

  .alert-dialog p {
    margin: 8px 0 0;
    color: #4b5563;
    font-size: 14px;
    line-height: 1.5;
  }

  .dialog-actions {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }

  .dialog-btn {
    border-radius: 8px;
    border: none;
    background: #f3f4f6;
    color: #374151;
    padding: 8px 14px;
    font-size: 13px;
    cursor: pointer;
  }

  .dialog-btn:hover:not(:disabled) {
    background: #e5e7eb;
  }

  .dialog-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .dialog-btn.danger {
    background: #dc2626;
    color: #fff;
  }

  .dialog-btn.danger:hover:not(:disabled) {
    background: #b91c1c;
  }
</style>
