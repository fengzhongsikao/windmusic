<!--
  最近播放：数据来自本地 recent.json，开始播放时自动记录。
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import { Clock, Music } from '@lucide/svelte';
  import {
    clearRecentHistory,
    fetchRecentSongs,
    onRecentChanged,
    removeRecentSong,
    type RecentSong,
  } from '@/lib/wailsPlayer';
  import { formatPlayedAt } from '@/lib/recentTime';
  import PlayAllButton from '@/components/PlayAllButton.svelte';
  import { player, playAllTracks, togglePlayByTrack } from '@/stores/player.svelte';
  import type { PlayerTrack } from '@/stores/player.svelte';

  let loading = $state(false);
  let error = $state('');
  let recentSongs = $state<RecentSong[]>([]);
  let brokenCovers = $state<Record<string, true>>({});
  let editMode = $state(false);
  let selectedIds = $state<Record<string, true>>({});
  let showClearDialog = $state(false);
  let showDeleteDialog = $state(false);
  let clearing = $state(false);
  let deleting = $state(false);

  function songKey(song: RecentSong): string {
    return [song.id, song.sourceId, song.platform, song.metaJson].join('|');
  }

  function recentToPlayerTrack(song: RecentSong): PlayerTrack {
    return {
      id: song.id,
      title: song.title,
      artist: song.artist,
      album: song.album ?? '',
      duration: song.duration?.trim() || '—',
      coverUrl: song.coverUrl?.trim() || undefined,
      playback:
        song.sourceId || song.platform || song.metaJson
          ? {
              sourceId: song.sourceId ?? '',
              platform: song.platform ?? '',
              metaJson: song.metaJson ?? '',
            }
          : undefined,
    };
  }

  function playSong(song: RecentSong) {
    if (editMode) return;
    togglePlayByTrack(recentToPlayerTrack(song));
  }

  function playAll() {
    if (editMode) return;
    playAllTracks(recentSongs.map(recentToPlayerTrack));
  }

  function handleCoverError(url: string) {
    if (!url || brokenCovers[url]) return;
    brokenCovers = { ...brokenCovers, [url]: true };
  }

  function toggleSelect(song: RecentSong, selected: boolean) {
    const id = songKey(song);
    if (selected) {
      selectedIds = { ...selectedIds, [id]: true };
      return;
    }
    const next = { ...selectedIds };
    delete next[id];
    selectedIds = next;
  }

  function startEdit() {
    editMode = true;
    selectedIds = {};
  }

  function cancelEdit() {
    editMode = false;
    selectedIds = {};
    showDeleteDialog = false;
  }

  const selectedCount = $derived(Object.keys(selectedIds).length);

  async function confirmDeleteSelected() {
    if (selectedCount === 0 || deleting) return;
    deleting = true;
    try {
      const targets = recentSongs.filter((song) => selectedIds[songKey(song)]);
      for (const song of targets) {
        await removeRecentSong(song);
      }
      showDeleteDialog = false;
      cancelEdit();
      await loadRecent();
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    } finally {
      deleting = false;
    }
  }

  async function confirmClearAll() {
    if (clearing) return;
    clearing = true;
    try {
      await clearRecentHistory();
      showClearDialog = false;
      cancelEdit();
      await loadRecent();
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    } finally {
      clearing = false;
    }
  }

  async function loadRecent() {
    loading = true;
    error = '';
    try {
      recentSongs = await fetchRecentSongs();
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
      recentSongs = [];
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    void loadRecent();
    return onRecentChanged(() => {
      void loadRecent();
    });
  });

</script>

<div class="recent-page">
  <div class="page-header">
    <div class="header-icon">
      <Clock size={32} />
    </div>
    <div class="header-info">
      <h2 class="section-title">最近播放</h2>
      <div class="song-count">{recentSongs.length} 首歌曲</div>
    </div>
    <div class="header-actions">
      {#if editMode}
        <button type="button" class="btn action-btn" onclick={cancelEdit}>取消</button>
        <button
          type="button"
          class="btn action-btn danger"
          disabled={selectedCount === 0}
          onclick={() => (showDeleteDialog = true)}
        >
          删除（{selectedCount}）
        </button>
      {:else}
        <PlayAllButton disabled={recentSongs.length === 0} onclick={playAll} />
        <button
          type="button"
          class="btn action-btn"
          disabled={recentSongs.length === 0}
          onclick={() => (showClearDialog = true)}
        >
          清空
        </button>
        <button type="button" class="btn action-btn" onclick={startEdit}>编辑</button>
      {/if}
    </div>
  </div>

  {#if error}
    <div class="feedback error">{error}</div>
  {:else if loading}
    <div class="feedback">正在加载最近播放…</div>
  {:else if recentSongs.length === 0}
    <div class="feedback">还没有播放记录，去搜索或发现页听首歌吧。</div>
  {:else}
    <div class="song-list">
      {#each recentSongs as song, index (songKey(song))}
        {@const cover = song.coverUrl?.trim() ?? ''}
        {@const hasCover = cover !== '' && !brokenCovers[cover]}
        {@const active =
          String(player.currentSong.id) === String(song.id) &&
          (player.currentSong.playback?.metaJson ?? '') === (song.metaJson ?? '')}
        <button
          type="button"
          class="song-row"
          class:active={active}
          class:selection-mode={editMode}
          onclick={() => {
            if (editMode) {
              const id = songKey(song);
              toggleSelect(song, !selectedIds[id]);
              return;
            }
            playSong(song);
          }}
        >
          {#if editMode}
            <span class="col-select">
              <input
                type="checkbox"
                checked={Boolean(selectedIds[songKey(song)])}
                onclick={(e) => e.stopPropagation()}
                onchange={(e) => toggleSelect(song, e.currentTarget.checked)}
              />
            </span>
          {/if}
          <span class="col-index">{index + 1}</span>
          <span class="col-title">
            <span class="song-cover">
              {#if hasCover}
                <img src={cover} alt="" onerror={() => handleCoverError(cover)} />
              {:else}
                <Music size={14} />
              {/if}
            </span>
            <span class="song-name">{song.title}</span>
            {#if active && player.isPlaying}
              <span class="playing-badge">播放中</span>
            {/if}
          </span>
          <span class="col-artist">{song.artist}</span>
          <span class="col-time">{formatPlayedAt(song.playedAt)}</span>
          <span class="col-duration">{song.duration?.trim() || '—'}</span>
        </button>
      {/each}
    </div>
  {/if}
</div>

{#if showClearDialog}
  <div class="dialog-backdrop" role="presentation">
    <div class="alert-dialog" role="alertdialog" aria-modal="true" aria-labelledby="clear-recent-title">
      <h3 id="clear-recent-title">清空最近播放？</h3>
      <p>将删除全部播放记录，此操作不可恢复。</p>
      <div class="dialog-actions">
        <button type="button" class="btn action-btn" onclick={() => (showClearDialog = false)}>取消</button>
        <button type="button" class="btn action-btn danger" disabled={clearing} onclick={confirmClearAll}>
          {clearing ? '清空中…' : '确定清空'}
        </button>
      </div>
    </div>
  </div>
{/if}

{#if showDeleteDialog}
  <div class="dialog-backdrop" role="presentation">
    <div class="alert-dialog" role="alertdialog" aria-modal="true" aria-labelledby="delete-recent-title">
      <h3 id="delete-recent-title">确定删除吗？</h3>
      <p>将从最近播放中删除已选择的歌曲。</p>
      <div class="dialog-actions">
        <button type="button" class="btn action-btn" onclick={() => (showDeleteDialog = false)}>取消</button>
        <button
          type="button"
          class="btn action-btn danger"
          disabled={deleting}
          onclick={confirmDeleteSelected}
        >
          {deleting ? '删除中…' : '确定删除'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .recent-page {
    padding: 0;
  }

  .page-header {
    display: flex;
    align-items: center;
    gap: 20px;
    margin-bottom: 24px;
    padding: 24px;
    background: linear-gradient(135deg, #667eea, #764ba2);
    border-radius: 16px;
    color: #fff;
  }

  .header-actions {
    margin-left: auto;
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }

  .action-btn {
    border-radius: 999px;
    border: none;
    background: rgba(255, 255, 255, 0.2);
    color: #fff;
    padding: 6px 14px;
    font-size: 13px;
    cursor: pointer;
  }

  .action-btn:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.3);
  }

  .action-btn:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }

  .action-btn.danger {
    background: rgba(220, 38, 38, 0.85);
  }

  .header-icon {
    width: 64px;
    height: 64px;
    background: rgba(255, 255, 255, 0.2);
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .section-title {
    font-size: 24px;
    font-weight: 700;
    margin: 0 0 4px;
  }

  .song-count {
    font-size: 14px;
    opacity: 0.8;
  }

  .feedback {
    padding: 64px 24px;
    text-align: center;
    color: #999;
    border-radius: 12px;
    background: #fafafa;
  }

  .feedback.error {
    color: #dc2626;
  }

  .song-list {
    border-radius: 12px;
    overflow: hidden;
    background: #fafafa;
  }

  .song-row {
    display: grid;
    grid-template-columns: 50px 1fr 140px 120px 80px;
    padding: 12px 16px;
    border: none;
    background: transparent;
    color: #666;
    font-size: 14px;
    cursor: pointer;
    transition: all 0.15s ease;
    text-align: left;
    width: 100%;
    align-items: center;
  }

  .song-row.selection-mode {
    grid-template-columns: 2rem 50px 1fr 120px 100px 80px;
  }

  .song-row:nth-child(even) {
    background: rgba(0, 0, 0, 0.01);
  }

  .song-row:hover {
    background: rgba(0, 0, 0, 0.04);
  }

  .song-row.active .song-name {
    color: #667eea;
  }

  .col-select {
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }

  .col-select input[type='checkbox'] {
    width: 16px;
    height: 16px;
    margin: 0;
    accent-color: #667eea;
    cursor: pointer;
  }

  .col-index {
    font-size: 13px;
    color: #ccc;
    text-align: center;
  }

  .col-title {
    display: flex;
    align-items: center;
    gap: 12px;
    overflow: hidden;
    min-width: 0;
  }

  .song-cover {
    width: 36px;
    height: 36px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #f0f0f0;
    border-radius: 6px;
    flex-shrink: 0;
    color: #999;
    overflow: hidden;
  }

  .song-cover img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .song-name {
    font-weight: 500;
    color: #333;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .playing-badge {
    flex-shrink: 0;
    font-size: 11px;
    color: #667eea;
    background: rgba(102, 126, 234, 0.12);
    padding: 2px 6px;
    border-radius: 4px;
  }

  .col-artist {
    color: #666;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .col-time {
    color: #999;
    font-size: 13px;
  }

  .col-duration {
    color: #ccc;
    text-align: right;
    font-size: 13px;
  }

  .dialog-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.35);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 50;
    padding: 16px;
  }

  .alert-dialog {
    width: min(420px, 100%);
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
  }

  .dialog-actions {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }

  .dialog-actions .action-btn {
    background: #f3f4f6;
    color: #374151;
  }

  .dialog-actions .action-btn.danger {
    background: #dc2626;
    color: #fff;
  }
</style>
