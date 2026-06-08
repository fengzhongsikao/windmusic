<!--
  底部播放条：播放控制、进度、音量（当前为 UI 占位，待接入真实播放器状态）。
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import { Music, Heart, Shuffle, SkipBack, Play, Pause, SkipForward, Repeat, Repeat1, MicVocal, Volume2, Volume1, Volume, VolumeX } from '@lucide/svelte';
  import {
    player,
    togglePlayerPlayback,
    openImmersiveView,
    playNextTrack,
    playPreviousTrack,
    toggleShuffleMode,
    cycleRepeatMode,
    repeatModeLabel,
    setPlayerVolume,
    togglePlayerMuted,
  } from '@/stores/playback/player.svelte';
  import PlayQueuePanel from '@/components/player/PlayQueuePanel.svelte';
  import AddToPlaylistMenu from '@/components/playlist/AddToPlaylistMenu.svelte';
  import { addTrackFavorite, checkTrackFavorite, fetchCoverUrl, onFavoritesChanged, removeTrackFavorite, toFavoriteSong } from '@/lib/wails/wailsPlayer';
  import { sameFavoriteSong } from '@/lib/library/favoriteSong';
  import {
    coverLookupKeys,
    fetchLocalSongCovers,
    localPathFromPlayerTrack,
    resolveCoverFromMaps,
  } from '@/lib/library/localMusic';
  import { favoritesState } from '@/stores/library/favorites.svelte';
  import { localLibrary } from '@/stores/library/localLibrary.svelte';
  import defaultCover from '@/assets/images/default.jpg';
  import {
    audioCurrentTime,
    audioDuration,
    audioLoading,
    seekAudio,
    setAudioVolume,
  } from '@/stores/playback/audioEngine';
  import VolumeSlider from '@/components/player/VolumeSlider.svelte';
  import ProgressSlider from '@/components/player/ProgressSlider.svelte';

  /** 仅打开详情页时切换沉浸式配色，单纯播放不变 */
  let barImmersive = $derived(player.viewMode === 'immersive');
  let barCoverByPath = $state<Record<string, string>>({});
  let displayedCover = $state('');
  const hasBarCover = $derived(displayedCover !== '' && displayedCover !== defaultCover);

  let currentTime = $derived($audioCurrentTime);
  let duration = $derived($audioDuration);
  let loadingAudio = $derived($audioLoading);
  let volume = $derived(player.volume);
  let isMuted = $derived(player.isMuted);
  let volumeDisplay = $derived(`${isMuted ? 0 : volume}%`);
  let isShuffled = $derived(player.isShuffled);
  let repeatMode = $derived(player.repeatMode);
  let repeatTitle = $derived(repeatModeLabel(repeatMode));
  let shuffleTitle = $derived(isShuffled ? '随机播放已开启' : '随机播放');
  let isFavorite = $state(false);
  let favoritePending = $state(false);
  let favoriteCheckToken = 0;
  let scrubTime = $state<number | null>(null);

  const displayTime = $derived(scrubTime ?? currentTime);

  function togglePlay() {
    togglePlayerPlayback();
  }

  function toggleMute() {
    togglePlayerMuted();
  }

  function toggleShuffle() {
    toggleShuffleMode();
  }

  function openSongDetail() {
    openImmersiveView();
  }

  function toggleRepeat() {
    cycleRepeatMode();
  }

  async function refreshFavoriteState(track = player.currentSong) {
    const token = ++favoriteCheckToken;
    try {
      const liked = await checkTrackFavorite(track);
      if (token === favoriteCheckToken) {
        isFavorite = liked;
      }
    } catch (error) {
      if (token === favoriteCheckToken) {
        isFavorite = false;
      }
      console.error('[播放器] 查询收藏状态失败', error);
    }
  }

  async function toggleFavorite() {
    if (favoritePending) return;
    favoritePending = true;
    const track = player.currentSong;
    try {
      if (isFavorite) {
        await removeTrackFavorite(track);
        isFavorite = false;
      } else {
        await addTrackFavorite(track);
        isFavorite = true;
      }
    } catch (error) {
      console.error('[播放器] 更新收藏失败', error);
    } finally {
      favoritePending = false;
    }
  }

  function formatTime(seconds: number): string {
    const min = Math.floor(seconds / 60);
    const sec = Math.floor(seconds % 60);
    return `${min}:${sec.toString().padStart(2, '0')}`;
  }

  function handleProgressInput(seconds: number) {
    scrubTime = seconds;
  }

  function handleProgressChange(seconds: number) {
    scrubTime = null;
    if (duration <= 0) return;
    seekAudio(seconds);
  }

  $effect(() => {
    setAudioVolume(player.volume, player.isMuted);
  });

  $effect(() => {
    const track = player.currentSong;
    if (favoritesState.loaded) {
      isFavorite = favoritesState.items.some((item) =>
        sameFavoriteSong(item, toFavoriteSong(track)),
      );
      return;
    }
    void refreshFavoriteState(track);
  });

  $effect(() => {
    player.currentSong.id;
    scrubTime = null;
  });

  onMount(() =>
    onFavoritesChanged(() => {
      void refreshFavoriteState(player.currentSong);
    })
  );

  $effect(() => {
    const track = player.currentSong;
    const path = localPathFromPlayerTrack(track);

    if (path && !barCoverByPath[path] && !localLibrary.coverByPath[path]) {
      void fetchLocalSongCovers([path]).then((batch) => {
        let added = false;
        for (const [filePath, key] of Object.entries(batch.paths)) {
          const cover = batch.covers[key]?.trim();
          if (cover && !barCoverByPath[filePath]) {
            barCoverByPath[filePath] = cover;
            added = true;
          }
        }
        if (added) {
          barCoverByPath = { ...barCoverByPath };
        }
      });
    }

    const sync = resolveCoverFromMaps(track.coverUrl, coverLookupKeys(path, track.id), [
      barCoverByPath,
      localLibrary.coverByPath,
    ]);
    if (sync) {
      displayedCover = sync;
      return;
    }

    let cancelled = false;
    displayedCover = '';

    void (async () => {
      const resolved = await fetchCoverUrl(track);
      if (cancelled) return;
      if (!resolved || resolved === defaultCover) {
        return;
      }

      const img = new Image();
      img.onload = () => {
        if (!cancelled) {
          displayedCover = resolved;
          if (path && !barCoverByPath[path]) {
            barCoverByPath = { ...barCoverByPath, [path]: resolved };
          }
        }
      };
      img.onerror = () => {
        if (!cancelled) displayedCover = '';
      };
      img.src = resolved;
    })();

    return () => {
      cancelled = true;
    };
  });

  function handleVolumeInput(next: number) {
    setPlayerVolume(next, { persist: false });
  }

  function handleVolumeChange(next: number) {
    setPlayerVolume(next, { persist: true });
  }
</script>

<div class="player-bar" class:immersive={barImmersive}>
  {#if barImmersive}
    <div
      class="bar-bg-cover"
      style:background-image="url('{displayedCover || defaultCover}')"
      aria-hidden="true"
    ></div>
    <div class="bar-bg-scrim" aria-hidden="true"></div>
  {/if}

  <div class="bar-inner">
  <div class="song-info">
    <button
      type="button"
      class="song-cover"
      onclick={openSongDetail}
      title="查看歌曲详情与歌词"
      aria-label="查看歌曲详情与歌词"
    >
      {#if hasBarCover}
        <img src={displayedCover} alt="" class="song-cover-img" />
      {:else}
        <Music size={24} />
      {/if}
    </button>
    <div class="song-details">
      <button
        type="button"
        class="song-title song-title-btn"
        onclick={openSongDetail}
        title="查看歌曲详情与歌词"
      >
        {player.currentSong.title}
      </button>
      <div class="song-artist">{player.currentSong.artist}</div>
    </div>
    <div class="song-actions">
      <button
        type="button"
        class="like-btn"
        class:active={isFavorite}
        disabled={favoritePending}
        onclick={toggleFavorite}
        title={isFavorite ? '取消喜欢' : '喜欢'}
        aria-label={isFavorite ? '取消喜欢' : '喜欢'}
      >
        <Heart size={18} />
      </button>
      <AddToPlaylistMenu
        track={player.currentSong}
        immersive={barImmersive}
        triggerClass="like-btn"
      />
    </div>
  </div>

  <div class="player-controls">
    <div class="control-buttons">
      <button
        class="ctrl-btn small"
        class:active={isShuffled}
        onclick={toggleShuffle}
        title={shuffleTitle}
        aria-pressed={isShuffled}
      >
        <Shuffle size={16} />
      </button>
      <button type="button" class="ctrl-btn" title="上一首" onclick={() => playPreviousTrack()}>
        <SkipBack size={18} />
      </button>
      <button
        type="button"
        class="ctrl-btn play-btn"
        onclick={togglePlay}
        disabled={loadingAudio}
        title={loadingAudio ? '加载中…' : player.isPlaying ? '暂停' : '播放'}
      >
        {#if loadingAudio}
          <span class="loading-dot" aria-hidden="true"></span>
        {:else if player.isPlaying}
          <Pause size={18} />
        {:else}
          <Play size={18} />
        {/if}
      </button>
      <button type="button" class="ctrl-btn" title="下一首" onclick={() => playNextTrack()}>
        <SkipForward size={18} />
      </button>
      <button
        class="ctrl-btn small"
        class:active={repeatMode !== 'off'}
        onclick={toggleRepeat}
        title={repeatTitle}
        aria-pressed={repeatMode !== 'off'}
      >
        {#if repeatMode === 'one'}
          <Repeat1 size={16} />
        {:else}
          <Repeat size={16} />
        {/if}
      </button>
    </div>

    <div class="progress-section">
      <span class="time">{formatTime(displayTime)}</span>
      <ProgressSlider
        value={displayTime}
        max={duration > 0 ? duration : 1}
        step={0.1}
        disabled={duration <= 0}
        ariaLabel="播放进度"
        oninput={handleProgressInput}
        onchange={handleProgressChange}
      />
      <span class="time">{formatTime(duration)}</span>
    </div>
  </div>

  <div class="extra-controls">
    <PlayQueuePanel immersive={barImmersive} />
    <div class="volume-control">
      <button type="button" class="ctrl-btn small" onclick={toggleMute} title={isMuted ? '取消静音' : '静音'}>
        {#if isMuted}
          <VolumeX size={16} />
        {:else if volume > 50}
          <Volume2 size={16} />
        {:else if volume > 0}
          <Volume1 size={16} />
        {:else}
          <Volume size={16} />
        {/if}
      </button>
      <VolumeSlider
        value={volume}
        muted={isMuted}
        playing={player.isPlaying}
        oninput={handleVolumeInput}
        onchange={handleVolumeChange}
      />
      <span class="volume-percent" aria-label={`当前音量 ${volumeDisplay}`}>{volumeDisplay}</span>
    </div>
  </div>
  </div>
</div>

<style>
  .player-bar {
    position: relative;
    height: 80px;
    cursor: default;
    background: #fafafa;
    border-top: 1px solid #eee;
    color: #333;
    user-select: none;
    overflow: hidden;
    transition:
      background 0.55s cubic-bezier(0.4, 0, 0.2, 1),
      border-color 0.55s cubic-bezier(0.4, 0, 0.2, 1),
      color 0.45s ease;
  }

  .bar-inner {
    position: relative;
    z-index: 1;
    height: 100%;
    display: grid;
    grid-template-columns: 280px 1fr 220px;
    align-items: center;
    padding: 0 20px;
  }

  .bar-bg-cover {
    position: absolute;
    inset: -40% -10% 0;
    background-color: #16100d;
    background-size: cover;
    background-position: center top;
    background-repeat: no-repeat;
    filter: blur(48px) saturate(1.2) brightness(0.55);
    pointer-events: none;
    transition: opacity 0.55s ease;
  }

  .bar-bg-scrim {
    position: absolute;
    inset: 0;
    background-color: rgba(22, 16, 13, 0.92);
    backdrop-filter: blur(24px) saturate(1.1);
    -webkit-backdrop-filter: blur(24px) saturate(1.1);
    pointer-events: none;
  }

  .player-bar.immersive {
    border-top: none;
    color: rgba(255, 255, 255, 0.92);
  }

  .player-bar.immersive .song-cover {
    background: rgba(255, 255, 255, 0.1);
    color: rgba(255, 255, 255, 0.6);
  }

  .player-bar.immersive .song-title-btn:hover {
    color: #fff;
  }

  .player-bar.immersive .song-artist {
    color: rgba(255, 255, 255, 0.48);
  }

  .player-bar.immersive .like-btn {
    color: rgba(255, 255, 255, 0.4);
  }

  .player-bar.immersive .ctrl-btn {
    color: rgba(255, 255, 255, 0.72);
  }

  .player-bar.immersive .ctrl-btn:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.1);
  }

  .player-bar.immersive .play-btn {
    background: rgba(255, 255, 255, 0.16);
    color: #fff;
  }

  .player-bar.immersive .play-btn:hover {
    background: rgba(255, 255, 255, 0.24);
  }

  .player-bar.immersive .time {
    color: rgba(255, 255, 255, 0.45);
  }

  .song-info {
    display: flex;
    align-items: center;
    gap: 14px;
  }

  .song-cover {
    width: 50px;
    height: 50px;
    padding: 0;
    border: none;
    background: #f0f0f0;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    color: #999;
    cursor: pointer;
    overflow: hidden;
    transition: box-shadow 0.15s ease, transform 0.15s ease;
  }

  .song-cover:hover {
    box-shadow: 0 2px 8px rgba(102, 126, 234, 0.25);
    transform: scale(1.03);
  }

  .song-cover-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  .song-details {
    overflow: hidden;
  }

  .song-title-btn {
    display: block;
    width: 100%;
    padding: 0;
    border: none;
    background: none;
    font: inherit;
    text-align: left;
    cursor: pointer;
  }

  .song-title-btn:hover {
    color: #667eea;
    text-decoration: underline;
  }

  .song-title {
    font-size: 14px;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .song-artist {
    font-size: 12px;
    color: #999;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .song-actions {
    display: flex;
    align-items: center;
    gap: 2px;
    flex-shrink: 0;
  }

  .like-btn {
    background: none;
    border: none;
    color: #ccc;
    cursor: pointer;
    padding: 4px;
    transition: color 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  :global(.add-playlist-trigger:hover) {
    color: #667eea;
  }

  .like-btn:hover {
    color: #e74c3c;
  }

  .like-btn.active {
    color: #e74c3c;
  }

  .like-btn.active :global(svg) {
    fill: currentColor;
  }

  .player-controls {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
  }

  .control-buttons {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .ctrl-btn {
    width: 36px;
    height: 36px;
    border: none;
    background: transparent;
    color: #666;
    cursor: pointer;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s;
  }

  .ctrl-btn:hover {
    color: #333;
    background: rgba(0, 0, 0, 0.06);
  }

  .ctrl-btn.small {
    width: 32px;
    height: 32px;
  }

  .ctrl-btn.active {
    color: #667eea;
  }

  .play-btn {
    width: 42px;
    height: 42px;
    background: transparent;
    color: #666;
  }

  .play-btn:hover {
    background: rgba(0, 0, 0, 0.08);
    color: #333;
    transform: scale(1.05);
  }

  .play-btn:disabled {
    opacity: 0.7;
    cursor: wait;
    transform: none;
  }

  .loading-dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    border: 2px solid currentColor;
    border-top-color: transparent;
    animation: spin 0.8s linear infinite;
    display: inline-block;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .progress-section {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
    max-width: 500px;
  }

  .progress-section :global(.progress-slider) {
    flex: 1;
    min-width: 0;
  }

  .time {
    font-size: 11px;
    color: #999;
    min-width: 35px;
    text-align: center;
    font-variant-numeric: tabular-nums;
  }

  .extra-controls {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 4px;
  }

  .volume-control {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .volume-control :global(.skeleton-slider) {
    width: 88px;
    flex-shrink: 0;
  }

  .volume-percent {
    width: 38px;
    text-align: right;
    font-size: 11px;
    color: #888;
    font-variant-numeric: tabular-nums;
  }

  .player-bar.immersive .volume-percent {
    color: rgba(255, 255, 255, 0.62);
  }

</style>
