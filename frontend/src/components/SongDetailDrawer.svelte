<!--
  歌曲详情抽屉：封面模糊背景 + 全屏播放器（参考稿布局）。
-->
<script lang="ts">
  import { fly, fade } from 'svelte/transition';
  import { Settings, ChevronDown } from '@lucide/svelte';
  import SongDetail from '@/pages/song/SongDetail.svelte';
  import { get } from 'svelte/store';
  import { playerState } from '@/stores/player';
  import { songDetailOpen, closeSongDetailDrawer } from '@/stores/songDetailDrawer';
  import defaultCover from '@/assets/images/default.jpg';

  let coverSrc = $derived($playerState.currentTrack.coverUrl?.trim() || defaultCover);

  function handleBackdropClick() {
    closeSongDetailDrawer();
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape' && get(songDetailOpen)) {
      closeSongDetailDrawer();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if $songDetailOpen}
  <div class="drawer-root" role="presentation">
    <button
      type="button"
      class="drawer-backdrop"
      aria-label="关闭歌曲详情"
      transition:fade={{ duration: 300 }}
      onclick={handleBackdropClick}
    ></button>

    <div
      class="drawer-panel"
      role="dialog"
      aria-modal="true"
      aria-label="正在播放"
      in:fly={{ y: '100%', duration: 400, opacity: 1 }}
      out:fly={{ y: '100%', duration: 340, opacity: 1 }}
    >
      <div
        class="bg-blur"
        style="background-image: url({coverSrc})"
        aria-hidden="true"
      ></div>
      <div class="bg-overlay" aria-hidden="true"></div>

      <header class="drawer-top">
        <button
          type="button"
          class="close-handle"
          aria-label="关闭"
          onclick={closeSongDetailDrawer}
        >
          <ChevronDown size={20} strokeWidth={2} />
        </button>

        <button type="button" class="settings-btn" aria-label="界面设置">
          <Settings size={16} strokeWidth={1.75} />
          <span>界面设置</span>
          <ChevronDown size={14} class="settings-chevron" />
        </button>
      </header>

      <div class="drawer-body">
        <SongDetail />
      </div>
    </div>
  </div>
{/if}

<style>
  .drawer-root {
    position: fixed;
    inset: 0;
    z-index: 100;
  }

  .drawer-backdrop {
    position: fixed;
    inset: 0;
    border: none;
    padding: 0;
    background: rgba(8, 5, 4, 0.5);
    cursor: pointer;
  }

  .drawer-panel {
    position: fixed;
    inset: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: #1c1512;
    will-change: transform;
  }

  .bg-blur {
    position: absolute;
    inset: -25%;
    background-size: cover;
    background-position: center;
    filter: blur(90px) saturate(1.3) brightness(0.5);
    transform: scale(1.12);
    opacity: 0.9;
    pointer-events: none;
  }

  .bg-overlay {
    position: absolute;
    inset: 0;
    background:
      radial-gradient(ellipse 100% 80% at 30% 20%, rgba(100, 65, 45, 0.35), transparent 55%),
      radial-gradient(ellipse 90% 70% at 80% 80%, rgba(60, 45, 38, 0.4), transparent 50%),
      linear-gradient(180deg, rgba(35, 25, 20, 0.75) 0%, rgba(22, 16, 13, 0.92) 100%);
    pointer-events: none;
  }

  .drawer-top {
    position: relative;
    z-index: 2;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 22px 0;
  }

  .close-handle {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    border: none;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.08);
    color: rgba(255, 255, 255, 0.7);
    cursor: pointer;
    transition: background 0.15s ease;
  }

  .close-handle:hover {
    background: rgba(255, 255, 255, 0.14);
    color: #fff;
  }

  .settings-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    border: none;
    border-radius: 8px;
    background: transparent;
    color: rgba(255, 255, 255, 0.55);
    font-size: 0.8125rem;
    font-family: inherit;
    cursor: pointer;
    transition: color 0.15s ease, background 0.15s ease;
  }

  .settings-btn:hover {
    color: rgba(255, 255, 255, 0.85);
    background: rgba(255, 255, 255, 0.06);
  }

  .settings-btn :global(.settings-chevron) {
    opacity: 0.7;
  }

  .drawer-body {
    position: relative;
    z-index: 1;
    flex: 1;
    min-height: 0;
    overflow: hidden;
    padding: 4px 20px 16px;
  }

  @media (max-width: 768px) {
    .drawer-body {
      padding: 0 10px 12px;
    }
  }
</style>
