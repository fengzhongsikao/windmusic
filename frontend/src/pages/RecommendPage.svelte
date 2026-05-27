<script lang="ts">
  import { Card, Heading } from 'flowbite-svelte';
  import { ListMusic, Play } from '@lucide/svelte';

  const playlists = [
    { id: 1, title: '每日推荐', desc: '根据你的口味生成', count: 30, color: '#667eea' },
    { id: 2, title: '华语经典', desc: '那些年我们听过的歌', count: 50, color: '#f093fb' },
    { id: 3, title: '深夜情歌', desc: '适合夜晚聆听', count: 40, color: '#4facfe' },
    { id: 4, title: '运动节拍', desc: '让运动更有动力', count: 35, color: '#43e97b' },
    { id: 5, title: '工作专注', desc: '提升工作效率', count: 45, color: '#fa709a' },
    { id: 6, title: '午后时光', desc: '轻松惬意的旋律', count: 28, color: '#fee140' },
  ];
</script>

<div class="recommend-page">
  <Heading tag="h2" class="mb-6 text-2xl font-bold text-gray-800">推荐歌单</Heading>

  <div class="playlist-grid">
    {#each playlists as pl}
      <Card class="playlist-card overflow-hidden !p-0">
        <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
        <div class="playlist-inner" onclick={() => {}}>
          <div
            class="playlist-cover"
            style="background: linear-gradient(135deg, {pl.color}, {pl.color}88)"
          >
            <ListMusic size={32} />
            <button type="button" class="play-btn" aria-label="播放">
              <Play size={20} />
            </button>
          </div>
          <div class="playlist-info">
            <div class="playlist-name">{pl.title}</div>
            <div class="playlist-desc">{pl.desc}</div>
            <div class="playlist-count">{pl.count} 首</div>
          </div>
        </div>
      </Card>
    {/each}
  </div>
</div>

<style>
  .playlist-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    gap: 20px;
  }

  :global(.playlist-card) {
    border: none;
    box-shadow: none;
    transition: transform 0.2s, box-shadow 0.2s;
  }

  :global(.playlist-card:hover) {
    transform: translateY(-4px);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
  }

  .playlist-inner {
    cursor: pointer;
  }

  .playlist-cover {
    aspect-ratio: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: rgba(255, 255, 255, 0.8);
    position: relative;
  }

  .play-btn {
    position: absolute;
    bottom: 12px;
    right: 12px;
    width: 40px;
    height: 40px;
    border: none;
    background: rgba(255, 255, 255, 0.9);
    color: #333;
    border-radius: 50%;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    opacity: 0;
    transform: translateY(8px);
    transition: all 0.2s;
  }

  .playlist-inner:hover .play-btn {
    opacity: 1;
    transform: translateY(0);
  }

  .play-btn:hover {
    background: #fff;
    transform: scale(1.1);
  }

  .playlist-info {
    padding: 12px 16px;
  }

  .playlist-name {
    font-size: 14px;
    font-weight: 600;
    color: #333;
    margin-bottom: 4px;
  }

  .playlist-desc {
    font-size: 12px;
    color: #999;
    margin-bottom: 4px;
  }

  .playlist-count {
    font-size: 11px;
    color: #ccc;
  }
</style>
