<!--
  用户歌单详情：播放、编辑删除歌曲、删除歌单。
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { ListMusic } from '@lucide/svelte';
  import TrackList from '@/components/track/TrackList.svelte';
  import PlayAllButton from '@/components/track/PlayAllButton.svelte';
  import type { TrackItem } from '@/lib/playback/track';
  import type { FavoriteSong } from '@/lib/wails/wailsPlayer';
  import { favoriteSongKey } from '@/lib/wails/wailsPlayer';
  import {
    deleteUserPlaylist,
    fetchPlaylist,
    onPlaylistsChanged,
    playlistActionErrorMessage,
    removeSongFromPlaylist,
    type UserPlaylist,
  } from '@/lib/library/playlists';
  import {
    fetchLocalSongCovers,
    localPathFromStoredSong,
    storedSongToPlayerTrack,
  } from '@/lib/library/localMusic';
  import { localDefaultCover } from '@/lib/playback/playerDefaultCovers';
  import { player, playAllTracks, togglePlayByTrack } from '@/stores/playback/player.svelte';
  import { playlistsState, refreshPlaylistsFromBackend } from '@/stores/library/playlistsStore.svelte';
  import { error as toastError } from '@/stores/ui/toast';

  interface Props {
    params?: Record<string, string | null> | null;
  }

  let { params = null }: Props = $props();

  const playlistId = $derived(params?.id?.trim() ?? '');

  let loading = $state(true);
  let error = $state('');
  let playlist = $state<UserPlaylist | null>(null);
  let brokenCovers = $state<Record<string, true>>({});
  let coverByPath = $state<Record<string, string>>({});
  let editMode = $state(false);
  let selectedIds = $state<Record<string, true>>({});
  let showDeleteDialog = $state(false);
  let showDeletePlaylistDialog = $state(false);
  let deleting = $state(false);

  function resolvePlaylistCover(song: FavoriteSong): string | undefined {
    const stored = song.coverUrl?.trim();
    if (stored) {
      return stored;
    }
    const path = localPathFromStoredSong(song);
    if (path) {
      return coverByPath[path]?.trim() || localDefaultCover;
    }
    return localDefaultCover;
  }

  const tracks = $derived<TrackItem[]>(
    (playlist?.songs ?? []).map((song) => ({
      id: song.id,
      listKey: favoriteSongKey(song),
      title: song.title,
      artist: song.artist,
      album: song.album ?? '',
      duration: song.duration?.trim() || '—',
      coverUrl: resolvePlaylistCover(song),
    })),
  );

  let currentSongId = $derived(player.currentSong.id);

  function songToPlayerTrack(song: FavoriteSong) {
    return storedSongToPlayerTrack(song);
  }

  function playAll() {
    if (editMode || !playlist) return;
    playAllTracks(playlist.songs.map(songToPlayerTrack));
  }

  function playTrack(track: TrackItem) {
    if (editMode || !playlist) return;
    const song = playlist.songs.find((item) => favoriteSongKey(item) === track.listKey);
    togglePlayByTrack(song ? songToPlayerTrack(song) : {
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
    if (!playlist || deleting) return;
    deleting = true;
    try {
      const targets = playlist.songs.filter((song) => selectedIds[favoriteSongKey(song)]);
      for (const song of targets) {
        await removeSongFromPlaylist(playlist.id, song);
      }
      cancelEdit();
      await loadPlaylist(true);
    } catch (err) {
      toastError(playlistActionErrorMessage(err, '删除歌曲失败'));
    } finally {
      deleting = false;
      showDeleteDialog = false;
    }
  }

  async function confirmDeletePlaylist() {
    if (!playlist || deleting) return;
    deleting = true;
    try {
      await deleteUserPlaylist(playlist.id);
      push('/');
    } catch (err) {
      toastError(playlistActionErrorMessage(err, '删除歌单失败'));
    } finally {
      deleting = false;
      showDeletePlaylistDialog = false;
    }
  }

  async function loadPlaylist(force = false) {
    const showLoading = !playlist;
    if (showLoading) loading = true;
    error = '';
    try {
      const item = await fetchPlaylist(playlistId, { force });
      if (!item) {
        playlist = null;
        error = '歌单不存在或已被删除';
        return;
      }
      playlist = item;
    } catch (err) {
      playlist = null;
      error = playlistActionErrorMessage(err, '加载歌单失败');
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    void refreshPlaylistsFromBackend();
    return onPlaylistsChanged(() => {
      void loadPlaylist(true);
    });
  });

  $effect(() => {
    const id = playlistId;
    if (!id) return;
    void loadPlaylist(true);
  });

  $effect(() => {
    const id = playlistId;
    if (!id) return;
    playlistsState.items;
    const synced = playlistsState.items.find((item) => item.id === id);
    if (!synced) return;
    if (!playlist || playlist.id !== id || synced.songCount !== playlist.songCount) {
      playlist = synced;
    }
  });

  $effect(() => {
    const songs = playlist?.songs ?? [];
    const paths = [
      ...new Set(
        songs
          .map((song) => localPathFromStoredSong(song))
          .filter((path) => path && !coverByPath[path]),
      ),
    ];
    if (paths.length === 0) {
      return;
    }

    void fetchLocalSongCovers(paths).then((batch) => {
      let added = false;
      for (const [path, key] of Object.entries(batch.paths)) {
        const cover = batch.covers[key]?.trim();
        if (cover && !coverByPath[path]) {
          coverByPath[path] = cover;
          added = true;
        }
      }
      if (added) {
        coverByPath = { ...coverByPath };
      }
    });
  });
</script>

<div class="playlist-page">
  {#if loading && !playlist}
    <div class="feedback">正在加载歌单…</div>
  {:else if error}
    <div class="feedback error">{error}</div>
  {:else if playlist}
    <div class="page-header">
      <div class="header-icon">
        <ListMusic size={32} />
      </div>
      <div class="header-info">
        <h2 class="section-title">{playlist.name}</h2>
        <div class="song-count">{playlist.songCount} 首歌曲</div>
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
            移除（{selectedCount}）
          </button>
        {:else}
          <PlayAllButton disabled={tracks.length === 0} onclick={playAll} />
          <button type="button" class="btn action-btn" onclick={startEdit}>编辑</button>
          <button
            type="button"
            class="btn action-btn danger-outline"
            onclick={() => (showDeletePlaylistDialog = true)}
          >
            删除歌单
          </button>
        {/if}
      </div>
    </div>

    {#if tracks.length === 0}
      <div class="feedback">
        歌单还是空的。在播放条点击「添加到歌单」，把喜欢的歌曲加进来吧。
      </div>
    {:else}
      <TrackList
        {tracks}
        activeId={currentSongId}
        incremental
        initialBatch={100}
        batchSize={100}
        listId={playlistId}
        {brokenCovers}
        {coverByPath}
        onSelect={playTrack}
        onCoverError={handleCoverError}
        ariaLabel={`${playlist.name} 歌曲列表`}
        selectionMode={editMode}
        {selectedIds}
        onToggleSelect={toggleSelect}
      />
    {/if}
  {/if}
</div>

{#if showDeleteDialog}
  <div class="dialog-backdrop" role="presentation">
    <div class="alert-dialog" role="alertdialog" aria-modal="true" aria-labelledby="delete-song-title">
      <h3 id="delete-song-title">从歌单移除？</h3>
      <p>将从「{playlist?.name}」中移除已选择的歌曲。</p>
      <div class="dialog-actions">
        <button type="button" class="btn action-btn" onclick={() => (showDeleteDialog = false)}>取消</button>
        <button
          type="button"
          class="btn action-btn danger"
          disabled={deleting}
          onclick={confirmDeleteSelected}
        >
          {deleting ? '移除中…' : '确定移除'}
        </button>
      </div>
    </div>
  </div>
{/if}

{#if showDeletePlaylistDialog}
  <div class="dialog-backdrop" role="presentation">
    <div class="alert-dialog" role="alertdialog" aria-modal="true" aria-labelledby="delete-playlist-title">
      <h3 id="delete-playlist-title">删除歌单？</h3>
      <p>将永久删除「{playlist?.name}」及其中的全部歌曲，此操作不可恢复。</p>
      <div class="dialog-actions">
        <button type="button" class="btn action-btn" onclick={() => (showDeletePlaylistDialog = false)}>
          取消
        </button>
        <button
          type="button"
          class="btn action-btn danger"
          disabled={deleting}
          onclick={confirmDeletePlaylist}
        >
          {deleting ? '删除中…' : '确定删除'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .playlist-page {
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
    flex-wrap: wrap;
    justify-content: flex-end;
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

  .action-btn.danger-outline {
    background: transparent;
    border: 1px solid rgba(255, 255, 255, 0.45);
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
