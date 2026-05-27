<!--
  排行榜页：榜单列表与排名趋势（当前为演示数据，待接后端）。
-->
<script lang="ts">
  import { Trophy, Play, TrendingUp, TrendingDown, Minus } from '@lucide/svelte';

  const rankings = [
    { rank: 1, title: '晴天', artist: '周杰伦', trend: 'up', change: 2 },
    { rank: 2, title: '稻香', artist: '周杰伦', trend: 'down', change: 1 },
    { rank: 3, title: '夜曲', artist: '周杰伦', trend: 'same', change: 0 },
    { rank: 4, title: '起风了', artist: '买辣椒也用券', trend: 'up', change: 5 },
    { rank: 5, title: '光年之外', artist: '邓紫棋', trend: 'up', change: 3 },
    { rank: 6, title: '平凡之路', artist: '朴树', trend: 'down', change: 2 },
    { rank: 7, title: '孤勇者', artist: '陈奕迅', trend: 'up', change: 1 },
    { rank: 8, title: '漠河舞厅', artist: '柳爽', trend: 'same', change: 0 },
    { rank: 9, title: '错位时空', artist: '艾辰', trend: 'down', change: 4 },
    { rank: 10, title: '踏山河', artist: '是七叔呢', trend: 'up', change: 2 },
  ];

  function getTrendIcon(trend: string) {
    if (trend === 'up') return TrendingUp;
    if (trend === 'down') return TrendingDown;
    return Minus;
  }

  function getTrendColor(trend: string) {
    if (trend === 'up') return '#43e97b';
    if (trend === 'down') return '#e74c3c';
    return '#999';
  }
</script>

<div class="ranking-page">
  <h2 class="section-title">
    <Trophy size={28} />
    排行榜
  </h2>

  <div class="ranking-list">
    {#each rankings as item}
      <div class="ranking-item" class:top3={item.rank <= 3}>
        <div class="rank" class:gold={item.rank === 1} class:silver={item.rank === 2} class:bronze={item.rank === 3}>
          {item.rank}
        </div>
        <div class="song-info">
          <div class="song-title">{item.title}</div>
          <div class="song-artist">{item.artist}</div>
        </div>
        <div class="trend" style="color: {getTrendColor(item.trend)}">
          <svelte:component this={getTrendIcon(item.trend)} size={14} />
          {#if item.change > 0}
            <span>{item.change}</span>
          {/if}
        </div>
        <button class="play-btn">
          <Play size={16} />
        </button>
      </div>
    {/each}
  </div>
</div>

<style>
  .ranking-page {
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

  .ranking-list {
    background: #fafafa;
    border-radius: 12px;
    overflow: hidden;
  }

  .ranking-item {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 14px 20px;
    border-bottom: 1px solid #eee;
    transition: background 0.15s;
  }

  .ranking-item:last-child {
    border-bottom: none;
  }

  .ranking-item:hover {
    background: rgba(0, 0, 0, 0.02);
  }

  .ranking-item.top3 {
    background: rgba(102, 126, 234, 0.02);
  }

  .rank {
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 16px;
    font-weight: 700;
    color: #999;
    flex-shrink: 0;
  }

  .rank.gold {
    color: #ffd700;
  }

  .rank.silver {
    color: #c0c0c0;
  }

  .rank.bronze {
    color: #cd7f32;
  }

  .song-info {
    flex: 1;
    min-width: 0;
  }

  .song-title {
    font-size: 14px;
    font-weight: 500;
    color: #333;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .song-artist {
    font-size: 12px;
    color: #999;
  }

  .trend {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 12px;
    font-weight: 500;
    min-width: 40px;
    justify-content: center;
  }

  .play-btn {
    width: 32px;
    height: 32px;
    border: none;
    background: transparent;
    color: #999;
    cursor: pointer;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s;
  }

  .play-btn:hover {
    background: #667eea;
    color: #fff;
  }
</style>
