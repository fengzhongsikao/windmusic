<script lang="ts">
  import { Heart, Music, Play, Pause } from '@lucide/svelte';

  interface Song {
    id: number;
    title: string;
    artist: string;
    album: string;
    duration: string;
  }

  const favorites: Song[] = [
    { id: 1, title: '晴天', artist: '周杰伦', album: '叶惠美', duration: '4:29' },
    { id: 3, title: '夜曲', artist: '周杰伦', album: '十一月的萧邦', duration: '4:34' },
    { id: 5, title: '光年之外', artist: '邓紫棋', album: '光年之外', duration: '3:52' },
    { id: 7, title: '孤勇者', artist: '陈奕迅', album: '孤勇者', duration: '4:16' },
    { id: 10, title: '踏山河', artist: '是七叔呢', album: '踏山河', duration: '3:22' },
  ];

  let currentSong: Song | null = $state(null);
  let isPlaying = $state(false);

  function playSong(song: Song) {
    if (currentSong?.id === song.id) {
      isPlaying = !isPlaying;
    } else {
      currentSong = song;
      isPlaying = true;
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
  </div>

  <div class="song-list">
    {#each favorites as song, index}
      <button
        class="song-row"
        class:playing={currentSong?.id === song.id}
        onclick={() => playSong(song)}
      >
        <span class="col-index">
          {#if currentSong?.id === song.id && isPlaying}
            <span class="playing-indicator">
              <span></span><span></span><span></span>
            </span>
          {:else}
            {index + 1}
          {/if}
        </span>
        <span class="col-title">
          <span class="song-cover">
            {#if currentSong?.id === song.id && isPlaying}
              <Play size={14} />
            {:else}
              <Music size={14} />
            {/if}
          </span>
          <span class="song-name">{song.title}</span>
        </span>
        <span class="col-artist">{song.artist}</span>
        <span class="col-album">{song.album}</span>
        <span class="col-duration">{song.duration}</span>
      </button>
    {/each}
  </div>
</div>

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

  .song-list {
    border-radius: 12px;
    overflow: hidden;
    background: #fafafa;
  }

  .song-row {
    display: grid;
    grid-template-columns: 50px 1fr 140px 160px 80px;
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

  .song-row:nth-child(even) {
    background: rgba(0, 0, 0, 0.01);
  }

  .song-row:hover {
    background: rgba(0, 0, 0, 0.04);
  }

  .song-row.playing {
    background: rgba(102, 126, 234, 0.08);
    color: #667eea;
  }

  .col-index {
    font-size: 13px;
    color: #ccc;
    text-align: center;
  }

  .song-row.playing .col-index {
    color: #667eea;
  }

  .col-title {
    display: flex;
    align-items: center;
    gap: 12px;
    overflow: hidden;
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
  }

  .song-row.playing .song-cover {
    color: #667eea;
    background: rgba(102, 126, 234, 0.1);
  }

  .song-name {
    font-weight: 500;
    color: #333;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .col-artist {
    color: #666;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .col-album {
    color: #999;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .col-duration {
    color: #ccc;
    text-align: right;
    font-size: 13px;
  }

  .playing-indicator {
    display: inline-flex;
    align-items: flex-end;
    gap: 2px;
    height: 14px;
  }

  .playing-indicator span {
    width: 3px;
    background: #667eea;
    border-radius: 1px;
    animation: bounce 0.8s ease-in-out infinite;
  }

  .playing-indicator span:nth-child(1) {
    height: 60%;
    animation-delay: 0s;
  }

  .playing-indicator span:nth-child(2) {
    height: 100%;
    animation-delay: 0.2s;
  }

  .playing-indicator span:nth-child(3) {
    height: 40%;
    animation-delay: 0.4s;
  }

  @keyframes bounce {
    0%, 100% { transform: scaleY(1); }
    50% { transform: scaleY(0.4); }
  }
</style>
