<script lang="ts">
  import { Clock, Music, Play } from '@lucide/svelte';

  interface Song {
    id: number;
    title: string;
    artist: string;
    playedAt: string;
    duration: string;
  }

  const recentSongs: Song[] = [
    { id: 1, title: '晴天', artist: '周杰伦', playedAt: '3 分钟前', duration: '4:29' },
    { id: 4, title: '起风了', artist: '买辣椒也用券', playedAt: '1 小时前', duration: '5:12' },
    { id: 6, title: '平凡之路', artist: '朴树', playedAt: '2 小时前', duration: '4:46' },
    { id: 2, title: '稻香', artist: '周杰伦', playedAt: '昨天', duration: '3:43' },
    { id: 8, title: '漠河舞厅', artist: '柳爽', playedAt: '昨天', duration: '4:38' },
    { id: 5, title: '光年之外', artist: '邓紫棋', playedAt: '2 天前', duration: '3:52' },
  ];
</script>

<div class="recent-page">
  <h2 class="section-title">
    <Clock size={28} />
    最近播放
  </h2>

  <div class="song-list">
    {#each recentSongs as song, index}
      <button class="song-row">
        <span class="col-index">{index + 1}</span>
        <span class="col-title">
          <span class="song-cover">
            <Music size={14} />
          </span>
          <span class="song-name">{song.title}</span>
        </span>
        <span class="col-artist">{song.artist}</span>
        <span class="col-time">{song.playedAt}</span>
        <span class="col-duration">{song.duration}</span>
      </button>
    {/each}
  </div>
</div>

<style>
  .recent-page {
    padding: 0;
  }

  .section-title {
    font-size: 24px;
    font-weight: 700;
    color: #333;
    margin-bottom: 24px;
    display: flex;
    align-items: center;
    gap: 12px;
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

  .song-row:nth-child(even) {
    background: rgba(0, 0, 0, 0.01);
  }

  .song-row:hover {
    background: rgba(0, 0, 0, 0.04);
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

  .col-time {
    color: #999;
    font-size: 13px;
  }

  .col-duration {
    color: #ccc;
    text-align: right;
    font-size: 13px;
  }
</style>
