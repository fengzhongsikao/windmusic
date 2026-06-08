<!--
  我喜欢的音乐：使用本地 favorites.json 数据，样式与首页列表保持一致。
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import { Heart } from '@lucide/svelte';
  import TrackList from '@/components/track/TrackList.svelte';
  import type { TrackItem } from '@/lib/track';
  import { favoriteSongKey, removeTrackFavorite } from '@/lib/wailsPlayer';
  import type { FavoriteSong } from '@/lib/wailsPlayer';
  import PlayAllButton from '@/components/track/PlayAllButton.svelte';
  import { storedSongToPlayerTrack } from '@/lib/localMusic';
  import { player, playAllTracks, togglePlayByTrack } from '@/stores/playback/player.svelte';
  import { favoritesState, refreshFavoritesFromBackend } from '@/stores/library/favorites.svelte';

  let brokenCovers = $state<Record<string, true>>({});
  let editMode = $state(false);
  let selectedIds = $state<Record<string, true>>({});
  let showDeleteDialog = $state(false);
  let deleting = $state(false);
  let error = $state('');

  const favorites = $derived(favoritesState.items);
  const loading = $derived(!favoritesState.loaded);

  const tracks = $derived<TrackItem[]>(
    favorites.map((song) => ({
      id: song.id,
      listKey: favoriteSongKey(song),
      title: song.title,
      artist: song.artist,
      album: song.album ?? '',
      duration: song.duration?.trim() || '—',
      coverUrl: song.coverUrl?.trim() || undefined,
    })),
  );

  let currentSongId = $derived(player.currentSong.id);

  onMount(() => {
    void refreshFavoritesFromBackend();
  });

  function favoriteToPlayerTrack(song: FavoriteSong) {
    return storedSongToPlayerTrack(song);
  }

  function playAll() {
    if (editMode) return;
    playAllTracks(favorites.map(favoriteToPlayerTrack));
  }

  function playTrack(track: TrackItem) {
    if (editMode) return;
    const song = favorites.find((item) => favoriteSongKey(item) === track.listKey);
    togglePlayByTrack(song ? favoriteToPlayerTrack(song) : {
      id: track.id,
      title: track.title,
      artist: track.artist,
      album: track.album,
      duration: track.duration,
      coverUrl: track.coverUrl,
    });
  }

  function handleCoverError(url: string) {
    if (!url || brokenCovers[url]) return;
    brokenCovers = { ...brokenCovers, [url]: true };
  }

  function toggleSelect(track: TrackItem, selected: boolean) {
    const key = track.listKey ?? String(track.id);
    if (selected) {
      selectedIds = { ...selectedIds, [key]: true };
      return;
    }
    const next = { ...selectedIds };
    delete next[key];
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
  const allSelected = $derived(tracks.length > 0 && selectedCount === tracks.length);

  function toggleSelectAll() {
    if (allSelected) {
      selectedIds = {};
      return;
    }
    const next: Record<string, true> = {};
    for (const track of tracks) {
      const key = track.listKey ?? String(track.id);
      next[key] = true;
    }
    selectedIds = next;
  }

  async function confirmDeleteSelected() {
    if (selectedCount === 0 || deleting) return;
    deleting = true;
    try {
      const targets = favorites.filter((song) => selectedIds[favoriteSongKey(song)]);
      for (const song of targets) {
        await removeTrackFavorite({
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
        });
      }
      showDeleteDialog = false;
      cancelEdit();
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    } finally {
      deleting = false;
    }
  }

</script>

<div class="favorites-page">
  <div class="page-header">
    <div class="header-icon">
      <Heart size={32} />
    </div>
    <div class="header-info">
      <h2 class="section-title">我喜欢的音乐</h2>
      <div class="song-count">{favorites.length} 首歌曲</div>
    </div>
    <div class="header-actions">
      {#if editMode}
        <button type="button" class="btn action-btn" onclick={cancelEdit}>取消</button>
        <button
          type="button"
          class="btn action-btn"
          disabled={tracks.length === 0}
          onclick={toggleSelectAll}
        >
          {allSelected ? '取消全选' : '全选'}
        </button>
        <button
          type="button"
          class="btn action-btn danger"
          disabled={selectedCount === 0}
          onclick={() => (showDeleteDialog = true)}
        >
          删除（{selectedCount}）
        </button>
      {:else}
        <PlayAllButton disabled={tracks.length === 0} onclick={playAll} />
        <button type="button" class="btn action-btn" onclick={startEdit}>编辑</button>
      {/if}
    </div>
  </div>

  {#if error}
    <div class="feedback error">{error}</div>
  {:else if loading}
    <div class="feedback">正在加载我喜欢的音乐…</div>
  {:else if tracks.length === 0}
    <div class="feedback">还没有收藏歌曲，去发现页或搜索页点亮爱心吧。</div>
  {:else}
    <TrackList
      {tracks}
      activeId={currentSongId}
      incremental
      initialBatch={100}
      batchSize={100}
      listId="favorites"
      {brokenCovers}
      onSelect={playTrack}
      onCoverError={handleCoverError}
      ariaLabel="我喜欢的音乐列表"
      selectionMode={editMode}
      {selectedIds}
      onToggleSelect={toggleSelect}
    />
  {/if}
</div>

{#if showDeleteDialog}
  <div class="dialog-backdrop" role="presentation">
    <div class="alert-dialog" role="alertdialog" aria-modal="true" aria-labelledby="delete-favorite-title">
      <h3 id="delete-favorite-title">确定删除吗？</h3>
      <p>将从我喜欢的音乐中删除已选择的歌曲。</p>
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
  .favorites-page {
    padding: 0;
  }

  .page-header {
    display: flex;
    align-items: center;
    gap: 20px;
    margin-bottom: 24px;
    padding: 24px;
    background: linear-gradient(135deg, #f093fb, #f5576c);
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
