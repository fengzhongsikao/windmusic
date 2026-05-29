<!--
  左侧导航：主导航、我的音乐、歌单列表；通过 hash 路由高亮当前项。
-->
<script lang="ts">
  import { Music, House, ListMusic, Trophy, Heart, Clock, FolderOpen, CirclePlus } from '@lucide/svelte';
  import { link } from 'svelte-spa-router';

  const menuItems = [
    { id: 'home', label: '首页', icon: House, path: '/' },
    // { id: 'recommend', label: '推荐歌单', icon: ListMusic, path: '/recommend' },
    // { id: 'ranking', label: '排行榜', icon: Trophy, path: '/ranking' },
  ];

  const libraryItems = [
    { id: 'favorites', label: '我喜欢的音乐', icon: Heart, path: '/favorites' },
    { id: 'recent', label: '最近播放', icon: Clock, path: '/recent' },
    { id: 'local', label: '本地音乐', icon: FolderOpen, path: '/local' },
  ];

  const playlists = [
    { id: 'pl1', label: '深夜放松' },
    { id: 'pl2', label: '工作专注' },
    { id: 'pl3', label: '运动节拍' },
  ];

  let currentPath = $state(window.location.hash.slice(1) || '/');

  function isActive(path: string): boolean {
    if (path === '/') {
      return currentPath === '/' || currentPath === '/discover' || currentPath === '';
    }
    return currentPath === path;
  }

  function handleNavigation(path: string) {
    currentPath = path;
  }

  function createPlaylist() {
    // TODO: 新建歌单
  }
</script>

<div class="sidebar">
  <div class="logo">
    <span class="logo-icon">
      <Music size={24} />
    </span>
    <span class="logo-text">WindMusic</span>
  </div>

  <div class="nav-section">
    <div class="section-title">在线音乐</div>
    {#each menuItems as item}
      <a
        href={item.path}
        use:link
        class="nav-item"
        class:active={isActive(item.path)}
        onclick={() => handleNavigation(item.path)}
      >
        <span class="nav-icon">
          <item.icon size={18} />
        </span>
        <span class="nav-label">{item.label}</span>
      </a>
    {/each}
  </div>

  <div class="nav-section">
    <div class="section-title">我的音乐</div>
    {#each libraryItems as item}
      <a
        href={item.path}
        use:link
        class="nav-item"
        class:active={isActive(item.path)}
        onclick={() => handleNavigation(item.path)}
      >
        <span class="nav-icon">
          <item.icon size={18} />
        </span>
        <span class="nav-label">{item.label}</span>
      </a>
    {/each}
  </div>

  <div class="nav-section">
    <div class="section-header">
      <div class="section-title">创建的歌单</div>
      <button
        type="button"
        class="add-playlist-btn"
        title="新建歌单"
        aria-label="新建歌单"
        onclick={createPlaylist}
      >
        <CirclePlus size={16} />
      </button>
    </div>
    {#each playlists as pl}
      <button
        class="nav-item playlist-item"
      >
        <span class="nav-icon">
          <Music size={16} />
        </span>
        <span class="nav-label">{pl.label}</span>
      </button>
    {/each}
  </div>
</div>

<style>
  .sidebar {
    width: 220px;
    height: 100%;
    background: #fafafa;
    color: #333;
    display: flex;
    flex-direction: column;
    overflow-y: auto;
    user-select: none;
    border-right: 1px solid #eee;
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 20px 20px 16px;
    font-size: 18px;
    font-weight: 700;
  }

  .logo-icon {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .logo-text {
    background: linear-gradient(135deg, #667eea, #764ba2);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .nav-section {
    padding: 8px 12px;
  }

  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px 4px;
  }

  .section-title {
    font-size: 11px;
    text-transform: uppercase;
    color: #999;
    padding: 8px 12px 4px;
    letter-spacing: 1px;
  }

  .section-header .section-title {
    padding: 0;
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 12px;
    width: 100%;
    padding: 10px 12px;
    border: none;
    background: transparent;
    color: #666;
    font-size: 14px;
    cursor: pointer;
    border-radius: 8px;
    transition: all 0.2s ease;
    text-align: left;
    text-decoration: none;
  }

  .nav-item:hover {
    background: rgba(0, 0, 0, 0.04);
    color: #333;
  }

  .nav-item.active {
    background: rgba(102, 126, 234, 0.1);
    color: #667eea;
  }

  .nav-icon {
    width: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .nav-label {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .playlist-item {
    font-size: 13px;
    padding: 8px 12px;
  }

  .add-playlist-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 2px;
    border: none;
    background: transparent;
    color: #999;
    cursor: pointer;
    border-radius: 4px;
    transition: all 0.2s ease;
  }

  .add-playlist-btn:hover {
    color: #667eea;
    background: rgba(102, 126, 234, 0.08);
  }
</style>
