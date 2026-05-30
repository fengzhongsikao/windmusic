<!--
  本地音乐：文件夹管理、扫描曲目、按文件夹 Tab 筛选、点击播放。
-->
<script lang="ts">
  import { Check, Folder, FolderOpen, LoaderCircle, Pencil, RefreshCw, Trash2, X } from '@lucide/svelte';
  import PlayAllButton from '@/components/track/PlayAllButton.svelte';
  import LocalFolderTrackPanel from '@/pages/library/LocalFolderTrackPanel.svelte';
  import type { TrackItem } from '@/lib/track';
  import {
    folderDisplayLabel,
    localSongToPlayerTrack,
    PickLocalMusicFolder,
    RemoveLocalMusicFolder,
    SetLocalFolderAlias,
  } from '@/lib/localMusic';
  import {
    LOCAL_ALL_TAB_ID,
    loadTracksIndex,
    localLibrary,
    scanLocalLibrary,
    setLocalActiveFolderTab,
  } from '@/stores/library/localLibrary.svelte';
  import { player, playAllTracks, togglePlayByTrack } from '@/stores/playback/player.svelte';
  import { audioLoading } from '@/stores/playback/audioEngine';
  import { error as toastError } from '@/stores/ui/toast';

  let pageError = $state('');
  let removingPath = $state('');
  let renamingPath = $state('');
  let renameDraft = $state('');
  let savingAlias = $state(false);
  let activeFolderTab = $state(LOCAL_ALL_TAB_ID);

  const folderAliases = $derived(localLibrary.folderAliases);
  const tracksIndexLoading = $derived(localLibrary.tracksIndexLoading);

  const folders = $derived(localLibrary.folders);
  const scanning = $derived(localLibrary.scanning);
  const loading = $derived(localLibrary.loading);

  const folderTabs = $derived.by(() => {
    const counts = localLibrary.folderCounts;
    const tabs: { id: string; label: string; count: number }[] = [
      { id: LOCAL_ALL_TAB_ID, label: '全部', count: counts[LOCAL_ALL_TAB_ID] ?? 0 },
    ];
    for (const folder of folders) {
      tabs.push({
        id: folder,
        label: folderDisplayLabel(folder, folderAliases),
        count: counts[folder] ?? 0,
      });
    }
    return tabs;
  });

  const activeTracks = $derived(localLibrary.tracksByTab[activeFolderTab] ?? []);
  const isLoadingAudio = $derived($audioLoading);

  $effect(() => {
    if (localLibrary.loaded) {
      void loadTracksIndex();
    }
  });

  function selectFolderTab(tabId: string) {
    if (activeFolderTab === tabId) {
      return;
    }
    activeFolderTab = tabId;
    setLocalActiveFolderTab(tabId);
  }

  function songCountForFolder(folderPath: string): number {
    return localLibrary.folderCounts[folderPath] ?? 0;
  }

  async function handleRescan() {
    pageError = '';
    try {
      await scanLocalLibrary();
      if (activeFolderTab !== LOCAL_ALL_TAB_ID && !folders.includes(activeFolderTab)) {
        activeFolderTab = LOCAL_ALL_TAB_ID;
      }
    } catch (err) {
      pageError = err instanceof Error ? err.message : String(err);
      toastError('扫描本地音乐失败');
    }
  }

  async function handlePickFolder() {
    pageError = '';
    try {
      const picked = await PickLocalMusicFolder();
      if (!picked?.trim()) {
        return;
      }
    } catch (err) {
      pageError = err instanceof Error ? err.message : String(err);
      toastError('添加音乐文件夹失败');
    }
  }

  function startRenameFolder(folderPath: string) {
    if (scanning || savingAlias) {
      return;
    }
    renamingPath = folderPath;
    renameDraft = folderDisplayLabel(folderPath, folderAliases);
  }

  function cancelRenameFolder() {
    renamingPath = '';
    renameDraft = '';
  }

  async function commitRenameFolder(folderPath: string) {
    if (savingAlias || renamingPath !== folderPath) {
      return;
    }
    const next = renameDraft.trim();
    const current = folderDisplayLabel(folderPath, folderAliases);
    if (next === current) {
      cancelRenameFolder();
      return;
    }
    savingAlias = true;
    pageError = '';
    try {
      await SetLocalFolderAlias(folderPath, next);
      cancelRenameFolder();
    } catch (err) {
      pageError = err instanceof Error ? err.message : String(err);
      toastError('保存文件夹名称失败');
    } finally {
      savingAlias = false;
    }
  }

  function handleRenameKeydown(event: KeyboardEvent, folderPath: string) {
    if (event.key === 'Enter') {
      event.preventDefault();
      void commitRenameFolder(folderPath);
    } else if (event.key === 'Escape') {
      event.preventDefault();
      cancelRenameFolder();
    }
  }

  async function handleRemoveFolder(folderPath: string) {
    if (removingPath) {
      return;
    }
    removingPath = folderPath;
    pageError = '';
    try {
      await RemoveLocalMusicFolder(folderPath);
      if (activeFolderTab === folderPath) {
        activeFolderTab = LOCAL_ALL_TAB_ID;
      }
    } catch (err) {
      pageError = err instanceof Error ? err.message : String(err);
      toastError('移除文件夹失败');
    } finally {
      removingPath = '';
    }
  }

  function handleSelect(track: TrackItem) {
    const song = localLibrary.songById.get(String(track.id));
    if (!song) {
      togglePlayByTrack({
        id: track.id,
        title: track.title,
        artist: track.artist,
        album: track.album,
        duration: track.duration,
        coverUrl: track.coverUrl,
      });
      return;
    }
    togglePlayByTrack(localSongToPlayerTrack(song, localLibrary.coverByPath[song.filePath]));
  }

  function handlePlayAll() {
    const queue = activeTracks.map((track) => {
      const song = localLibrary.songById.get(String(track.id));
      if (!song) {
        return {
          id: track.id,
          title: track.title,
          artist: track.artist,
          album: track.album,
          duration: track.duration,
          coverUrl: track.coverUrl,
        };
      }
      return localSongToPlayerTrack(song, localLibrary.coverByPath[song.filePath]);
    });
    playAllTracks(queue);
  }

</script>

<div class="local-page">
  <div class="page-header">
    <h2 class="section-title">本地音乐</h2>
    <div class="header-actions">
      {#if scanning}
        <span class="status-pill">
          <LoaderCircle size={14} class="spin" />
          扫描中…
        </span>
      {:else if loading}
        <span class="status-pill">加载中…</span>
      {:else if isLoadingAudio}
        <span class="status-pill">加载音频中…</span>
      {/if}
      <button
        type="button"
        class="btn-secondary"
        onclick={() => void handleRescan()}
        disabled={scanning || loading || folders.length === 0}
        title="重新扫描磁盘上的音乐文件"
      >
        <RefreshCw size={16} />
        刷新
      </button>
      <button type="button" class="btn-primary" onclick={handlePickFolder} disabled={scanning}>
        <FolderOpen size={16} />
        添加文件夹
      </button>
      {#if activeTracks.length > 0}
        <PlayAllButton onclick={handlePlayAll} disabled={scanning} />
      {/if}
    </div>
  </div>

  {#if pageError}
    <p class="page-error" role="alert">{pageError}</p>
  {/if}

  <div class="folder-section">
    <h3 class="sub-title">音乐文件夹</h3>
    {#if folders.length === 0}
      <p class="empty-hint">尚未添加文件夹，点击「添加文件夹」选择本地音乐目录。</p>
    {:else}
      <div class="folder-grid">
        {#each folders as folder (folder)}
          <div class="folder-card">
            <Folder size={28} />
            <div class="folder-info">
              {#if renamingPath === folder}
                <div class="folder-rename-row">
                  <input
                    type="text"
                    class="folder-rename-input"
                    bind:value={renameDraft}
                    maxlength={80}
                    aria-label="文件夹显示名称"
                    disabled={savingAlias}
                    onkeydown={(e) => handleRenameKeydown(e, folder)}
                  />
                  <button
                    type="button"
                    class="folder-action"
                    title="保存名称"
                    aria-label="保存名称"
                    disabled={savingAlias}
                    onclick={() => void commitRenameFolder(folder)}
                  >
                    <Check size={16} />
                  </button>
                  <button
                    type="button"
                    class="folder-action"
                    title="取消"
                    aria-label="取消"
                    disabled={savingAlias}
                    onclick={cancelRenameFolder}
                  >
                    <X size={16} />
                  </button>
                </div>
              {:else}
                <div class="folder-name" title={folder}>
                  {folderDisplayLabel(folder, folderAliases)}
                </div>
              {/if}
              <div class="folder-meta">
                <span class="folder-count">{songCountForFolder(folder)} 首</span>
                <span class="folder-path" title={folder}>{folder}</span>
              </div>
            </div>
            {#if renamingPath !== folder}
              <button
                type="button"
                class="folder-action"
                title="重命名显示名称"
                aria-label="重命名显示名称"
                disabled={removingPath === folder || scanning || savingAlias}
                onclick={() => startRenameFolder(folder)}
              >
                <Pencil size={16} />
              </button>
            {/if}
            <button
              type="button"
              class="folder-remove"
              title="移除此文件夹"
              aria-label="移除此文件夹"
              disabled={removingPath === folder || scanning || renamingPath === folder}
              onclick={() => void handleRemoveFolder(folder)}
            >
              <Trash2 size={16} />
            </button>
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <div class="song-section">
    <div class="song-section-header">
      <h3 class="sub-title">本地歌曲</h3>
      {#if folderTabs.length > 1}
        <div class="folder-tab-group" role="tablist" aria-label="按文件夹筛选">
          {#each folderTabs as tab (tab.id)}
            <button
              type="button"
              class="folder-tab"
              class:active={activeFolderTab === tab.id}
              role="tab"
              aria-selected={activeFolderTab === tab.id}
              onclick={() => selectFolderTab(tab.id)}
            >
              <span class="folder-tab-label">{tab.label}</span>
              <span class="folder-tab-count">{tab.count}</span>
            </button>
          {/each}
        </div>
      {/if}
    </div>

    {#if (scanning || loading || tracksIndexLoading) && !localLibrary.tracksIndexReady}
      <p class="empty-hint">
        {scanning ? '正在扫描音乐文件…' : tracksIndexLoading ? '正在加载歌曲…' : '正在加载…'}
      </p>
    {:else}
      <div class="folder-track-panels">
        {#each folderTabs as tab (tab.id)}
          <LocalFolderTrackPanel
            tabId={tab.id}
            tabLabel={tab.label}
            visible={activeFolderTab === tab.id}
            onSelect={handleSelect}
          />
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
  .local-page {
    padding: 0;
  }

  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 24px;
    flex-wrap: wrap;
  }

  .section-title {
    font-size: 24px;
    font-weight: 700;
    color: #333;
    margin: 0;
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  .status-pill {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    color: #667eea;
    background: rgba(102, 126, 234, 0.1);
    padding: 6px 10px;
    border-radius: 999px;
  }

  :global(.spin) {
    animation: spin 0.9s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .btn-primary,
  .btn-secondary {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 8px 14px;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    cursor: pointer;
  }

  .btn-primary {
    background: #667eea;
    color: #fff;
  }

  .btn-primary:hover:not(:disabled) {
    background: #5a6fd6;
  }

  .btn-secondary {
    background: #fff;
    color: #667eea;
    border: 1px solid #d1d5db;
  }

  .btn-secondary:hover:not(:disabled) {
    border-color: #667eea;
    background: #f8f9ff;
  }

  .btn-primary:disabled,
  .btn-secondary:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .page-error {
    color: #c0392b;
    font-size: 14px;
    margin: 0 0 16px;
  }

  .sub-title {
    font-size: 16px;
    font-weight: 600;
    color: #666;
    margin: 0;
  }

  .folder-section {
    margin-bottom: 32px;
  }

  .empty-hint {
    color: #999;
    font-size: 14px;
    margin: 0;
  }

  .folder-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 12px;
  }

  .folder-card {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 14px 16px;
    background: #f5f5f5;
    border-radius: 12px;
    color: #667eea;
  }

  .folder-info {
    flex: 1;
    min-width: 0;
    text-align: left;
  }

  .folder-name {
    font-size: 14px;
    font-weight: 600;
    color: #333;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .folder-rename-row {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
  }

  .folder-rename-input {
    flex: 1;
    min-width: 0;
    font-size: 14px;
    font-weight: 600;
    color: #333;
    border: 1px solid #c7d2fe;
    border-radius: 6px;
    padding: 4px 8px;
    background: #fff;
  }

  .folder-rename-input:focus {
    outline: none;
    border-color: #667eea;
    box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.15);
  }

  .folder-action {
    flex-shrink: 0;
    border: none;
    background: transparent;
    color: #999;
    cursor: pointer;
    padding: 4px;
    border-radius: 6px;
  }

  .folder-action:hover:not(:disabled) {
    color: #667eea;
    background: rgba(102, 126, 234, 0.08);
  }

  .folder-action:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .folder-meta {
    margin-top: 4px;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .folder-count {
    font-size: 12px;
    color: #999;
  }

  .folder-path {
    font-size: 11px;
    color: #bbb;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .folder-remove {
    flex-shrink: 0;
    border: none;
    background: transparent;
    color: #999;
    cursor: pointer;
    padding: 4px;
    border-radius: 6px;
  }

  .folder-remove:hover:not(:disabled) {
    color: #c0392b;
    background: rgba(192, 57, 43, 0.08);
  }

  .folder-remove:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .song-section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 16px;
    flex-wrap: wrap;
  }

  .folder-tab-group {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
    justify-content: flex-end;
    margin-left: auto;
  }

  .folder-tab {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    border: 1px solid #e5e7eb;
    border-radius: 999px;
    background: #fff;
    color: #666;
    font-size: 13px;
    cursor: pointer;
    transition:
      background 0.15s ease,
      color 0.15s ease,
      border-color 0.15s ease;
  }

  .folder-tab:hover {
    border-color: #c7d2fe;
    color: #667eea;
  }

  .folder-tab.active {
    background: #667eea;
    border-color: #667eea;
    color: #fff;
  }

  .folder-tab-count {
    font-size: 11px;
    opacity: 0.85;
  }

  .folder-tab.active .folder-tab-count {
    opacity: 0.95;
  }

  .folder-track-panels {
    position: relative;
    min-height: 120px;
  }

  .song-section :global(.track-list) {
    border-radius: 12px;
    overflow: hidden;
    background: #fafafa;
  }
</style>
