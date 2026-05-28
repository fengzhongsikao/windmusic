<!--
  底部播放条：播放控制、进度、音量（当前为 UI 占位，待接入真实播放器状态）。
-->
<script lang="ts">
  import { Music, Heart, Shuffle, SkipBack, Play, Pause, SkipForward, Repeat, Repeat1, MicVocal, ListMusic, Volume2, Volume1, Volume, VolumeX } from '@lucide/svelte';
  import { playerState, togglePlayerPlayback } from '@/stores/player';
  import { openSongDetailDrawer } from '@/stores/songDetailDrawer';
  import {
    audioCurrentTime,
    audioDuration,
    seekAudio,
    setAudioVolume,
  } from '@/stores/audioEngine';

  let isPlaying = $derived($playerState.isPlaying);
  let currentTime = $derived($audioCurrentTime);
  let duration = $derived($audioDuration);
  let volume = $state(75);
  let isMuted = $state(false);
  let isShuffled = $state(false);
  let repeatMode: 'off' | 'all' | 'one' = $state('off');

  let currentSong = $derived($playerState.currentTrack);

  function togglePlay() {
    togglePlayerPlayback();
  }

  function toggleMute() {
    isMuted = !isMuted;
  }

  function toggleShuffle() {
    isShuffled = !isShuffled;
  }

  function openSongDetail() {
    openSongDetailDrawer();
  }

  function toggleRepeat() {
    const modes: ('off' | 'all' | 'one')[] = ['off', 'all', 'one'];
    const currentIndex = modes.indexOf(repeatMode);
    repeatMode = modes[(currentIndex + 1) % modes.length];
  }

  function formatTime(seconds: number): string {
    const min = Math.floor(seconds / 60);
    const sec = Math.floor(seconds % 60);
    return `${min}:${sec.toString().padStart(2, '0')}`;
  }

  function handleProgressClick(e: MouseEvent) {
    if (duration <= 0) return;
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    const percent = (e.clientX - rect.left) / rect.width;
    seekAudio(percent * duration);
  }

  function handleProgressKeydown(e: KeyboardEvent) {
    if (duration <= 0) return;
    const step = 5;
    switch (e.key) {
      case 'ArrowLeft':
      case 'ArrowDown':
        e.preventDefault();
        seekAudio(Math.max(0, currentTime - step));
        break;
      case 'ArrowRight':
      case 'ArrowUp':
        e.preventDefault();
        seekAudio(Math.min(duration, currentTime + step));
        break;
      case 'Home':
        e.preventDefault();
        seekAudio(0);
        break;
      case 'End':
        e.preventDefault();
        seekAudio(duration);
        break;
    }
  }

  $effect(() => {
    setAudioVolume(volume, isMuted);
  });

  function handleVolumeClick(e: MouseEvent) {
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    volume = Math.round(((e.clientX - rect.left) / rect.width) * 100);
    isMuted = false;
  }

  function handleVolumeKeydown(e: KeyboardEvent) {
    const step = 5;
    switch (e.key) {
      case 'ArrowLeft':
      case 'ArrowDown':
        e.preventDefault();
        isMuted = false;
        volume = Math.max(0, volume - step);
        break;
      case 'ArrowRight':
      case 'ArrowUp':
        e.preventDefault();
        isMuted = false;
        volume = Math.min(100, volume + step);
        break;
      case 'Home':
        e.preventDefault();
        isMuted = false;
        volume = 0;
        break;
      case 'End':
        e.preventDefault();
        isMuted = false;
        volume = 100;
        break;
    }
  }
</script>

<div class="player-bar">
  <div class="song-info">
    <button
      type="button"
      class="song-cover"
      onclick={openSongDetail}
      title="查看歌曲详情与歌词"
      aria-label="查看歌曲详情与歌词"
    >
      {#if currentSong.coverUrl?.trim()}
        <img src={currentSong.coverUrl} alt="" class="song-cover-img" />
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
        {currentSong.title}
      </button>
      <div class="song-artist">{currentSong.artist}</div>
    </div>
    <button type="button" class="like-btn" title="喜欢">
      <Heart size={18} />
    </button>
  </div>

  <div class="player-controls">
    <div class="control-buttons">
      <button
        class="ctrl-btn small"
        class:active={isShuffled}
        onclick={toggleShuffle}
        title="随机播放"
      >
        <Shuffle size={16} />
      </button>
      <button type="button" class="ctrl-btn" title="上一首">
        <SkipBack size={18} />
      </button>
      <button type="button" class="ctrl-btn play-btn" onclick={togglePlay} title={isPlaying ? '暂停' : '播放'}>
        {#if isPlaying}
          <Pause size={18} />
        {:else}
          <Play size={18} />
        {/if}
      </button>
      <button type="button" class="ctrl-btn" title="下一首">
        <SkipForward size={18} />
      </button>
      <button
        class="ctrl-btn small"
        class:active={repeatMode !== 'off'}
        onclick={toggleRepeat}
        title="循环模式"
      >
        {#if repeatMode === 'one'}
          <Repeat1 size={16} />
        {:else}
          <Repeat size={16} />
        {/if}
      </button>
    </div>

    <div class="progress-section">
      <span class="time">{formatTime(currentTime)}</span>
      <div
        class="progress-bar"
        role="slider"
        tabindex="0"
        aria-label="播放进度"
        aria-valuemin={0}
        aria-valuemax={duration}
        aria-valuenow={currentTime}
        aria-valuetext={formatTime(currentTime)}
        onclick={handleProgressClick}
        onkeydown={handleProgressKeydown}
      >
        <div class="progress-bg">
          <div
            class="progress-fill"
            style="width: {duration > 0 ? (currentTime / duration) * 100 : 0}%"
          >
            <div class="progress-thumb"></div>
          </div>
        </div>
      </div>
      <span class="time">{formatTime(duration)}</span>
    </div>
  </div>

  <div class="extra-controls">
    <button
      type="button"
      class="ctrl-btn small"
      title="歌词"
      aria-label="打开歌词详情"
      onclick={openSongDetail}
    >
      <MicVocal size={16} />
    </button>
    <button type="button" class="ctrl-btn small" title="播放列表">
      <ListMusic size={16} />
    </button>
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
      <div
        class="volume-bar"
        role="slider"
        tabindex="0"
        aria-label="音量"
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={isMuted ? 0 : volume}
        aria-valuetext={isMuted ? '静音' : `${volume}%`}
        onclick={handleVolumeClick}
        onkeydown={handleVolumeKeydown}
      >
        <div class="volume-bg">
          <div
            class="volume-fill"
            style="width: {isMuted ? 0 : volume}%"
          ></div>
        </div>
      </div>
    </div>
  </div>
</div>

<style>
  .player-bar {
    height: 80px;
    background: #fafafa;
    border-top: 1px solid #eee;
    display: grid;
    grid-template-columns: 280px 1fr 220px;
    align-items: center;
    padding: 0 20px;
    color: #333;
    user-select: none;
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

  .like-btn:hover {
    color: #e74c3c;
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
    background: #667eea;
    color: #fff;
  }

  .play-btn:hover {
    background: #5a6fd6;
    transform: scale(1.05);
  }

  .progress-section {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
    max-width: 500px;
  }

  .time {
    font-size: 11px;
    color: #999;
    min-width: 35px;
    text-align: center;
    font-variant-numeric: tabular-nums;
  }

  .progress-bar {
    flex: 1;
    cursor: pointer;
    padding: 4px 0;
  }

  .progress-bg {
    height: 4px;
    background: #e0e0e0;
    border-radius: 2px;
    position: relative;
  }

  .progress-fill {
    height: 100%;
    background: #667eea;
    border-radius: 2px;
    position: relative;
    transition: width 0.1s linear;
  }

  .progress-thumb {
    width: 12px;
    height: 12px;
    background: #667eea;
    border-radius: 50%;
    position: absolute;
    right: -6px;
    top: -4px;
    opacity: 0;
    transition: opacity 0.15s;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.2);
  }

  .progress-bar:hover .progress-thumb {
    opacity: 1;
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

  .volume-bar {
    width: 80px;
    cursor: pointer;
    padding: 4px 0;
  }

  .volume-bg {
    height: 4px;
    background: #e0e0e0;
    border-radius: 2px;
  }

  .volume-fill {
    height: 100%;
    background: #667eea;
    border-radius: 2px;
    transition: width 0.1s;
  }
</style>
