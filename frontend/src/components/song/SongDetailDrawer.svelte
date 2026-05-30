<!--
  全屏详情抽屉：自带沉浸式模糊背景；底部 80px 留给全局 PlayerBar。
-->
<script lang="ts">
  import { fly } from 'svelte/transition';
  import { ChevronDown, Settings } from '@lucide/svelte';
  import SongDetail from '@/pages/song/SongDetail.svelte';
  import { player, closeImmersiveView } from '@/stores/playback/player.svelte';
  import { fetchCoverUrl } from '@/lib/wailsPlayer';
  import defaultCover from '@/assets/images/default.jpg';

  let coverSrc = $derived(player.currentSong.coverUrl?.trim() || defaultCover);
  let displayedCover = $state(defaultCover);

  $effect(() => {
    const target = coverSrc;
    let cancelled = false;

    void (async () => {
      const resolved =
        !target || target === defaultCover ? await fetchCoverUrl(player.currentSong) : target;
      if (cancelled) return;

      const img = new Image();
      img.onload = () => {
        if (!cancelled) displayedCover = resolved;
      };
      img.onerror = () => {
        if (!cancelled) displayedCover = defaultCover;
      };
      img.src = resolved;
    })();

    return () => {
      cancelled = true;
    };
  });

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape' && player.viewMode === 'immersive') {
      closeImmersiveView();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if player.viewMode === 'immersive'}
  <div class="detail-layer" role="presentation">
    <div
      class="detail-bg"
      style:background-image="url('{displayedCover}')"
      aria-hidden="true"
    ></div>
    <div class="detail-scrim" aria-hidden="true"></div>

    <div
      class="detail-panel"
      role="dialog"
      aria-modal="true"
      aria-label="正在播放"
      in:fly={{ y: '100%', duration: 120, opacity: 1 }}
      out:fly={{ y: '100%', duration: 90, opacity: 1 }}
    >
      <header class="detail-top">
        <button
          type="button"
          class="close-handle"
          aria-label="关闭"
          onclick={closeImmersiveView}
        >
          <ChevronDown size={20} strokeWidth={2} />
        </button>

        <button type="button" class="settings-btn" aria-label="界面设置">
          <Settings size={16} strokeWidth={1.75} />
          <span>界面设置</span>
          <ChevronDown size={14} class="settings-chevron" />
        </button>
      </header>

      <div class="detail-body">
        <SongDetail />
      </div>
    </div>
  </div>
{/if}

<style>
  .detail-layer {
    position: fixed;
    inset: 0;
    z-index: 100;
    overflow: hidden;
  }

  .detail-bg {
    position: absolute;
    inset: -25%;
    background-size: cover;
    background-position: center;
    filter: blur(90px) saturate(1.25) brightness(0.5);
    transform: scale(1.1);
    transition: opacity 0.55s ease;
  }

  .detail-scrim {
    position: absolute;
    inset: 0;
    background:
      radial-gradient(ellipse 90% 70% at 25% 20%, rgba(90, 58, 42, 0.4), transparent 55%),
      linear-gradient(180deg, rgba(24, 18, 14, 0.75) 0%, rgba(14, 10, 8, 0.94) 100%);
  }

  .detail-panel {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 80px;
    z-index: 1;
    display: flex;
    flex-direction: column;
    will-change: transform;
  }

  .detail-panel::after {
    content: '';
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    height: 2px;
    background: rgba(14, 10, 8, 0.98);
    pointer-events: none;
  }

  .detail-top {
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
    color: rgba(255, 255, 255, 0.75);
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
  }

  .settings-btn:hover {
    color: rgba(255, 255, 255, 0.85);
    background: rgba(255, 255, 255, 0.06);
  }

  .detail-body {
    flex: 1;
    min-height: 0;
    overflow: hidden;
    padding: 4px 20px 16px;
  }
</style>
