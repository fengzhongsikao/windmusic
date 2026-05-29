<!--
  当前播放队列：Skeleton Menu 下拉面板。
-->
<script lang="ts">
  import { ListMusic, Music, Play, Trash2 } from '@lucide/svelte';
  import { Menu, Portal } from '@skeletonlabs/skeleton-svelte';
  import {
    player,
    isCurrentTrack,
    playQueueTrack,
    removeQueueTrack,
    clearQueue,
    togglePlayerPlayback,
    type PlayerTrack,
  } from '@/stores/player.svelte';

  interface Props {
    immersive?: boolean;
  }

  let { immersive = false }: Props = $props();

  let brokenCovers = $state<Record<string, true>>({});

  const queueCount = $derived(player.queue.length);

  function trackKey(track: PlayerTrack, index: number): string {
    return [
      index,
      track.id,
      track.playback?.sourceId ?? '',
      track.playback?.platform ?? '',
      track.playback?.metaJson ?? '',
    ].join('|');
  }

  function handleCoverError(url: string) {
    if (!url || brokenCovers[url]) return;
    brokenCovers = { ...brokenCovers, [url]: true };
  }

  function playTrackAt(index: number) {
    const track = player.queue[index];
    if (!track) return;
    if (isCurrentTrack(track)) {
      togglePlayerPlayback();
      return;
    }
    playQueueTrack(index);
  }

  function handleSelect(details: { value: string }) {
    const { value } = details;
    if (value === 'clear') {
      clearQueue();
      return;
    }
    if (value.startsWith('play:')) {
      const index = Number(value.slice(5));
      if (Number.isFinite(index)) {
        playTrackAt(index);
      }
    }
  }
</script>

<Menu
  class="queue-menu-root"
  closeOnSelect={false}
  aria-label="播放列表"
  positioning={{ placement: 'top-end', gutter: 10 }}
  onSelect={handleSelect}
>
  <Menu.Trigger
    class="ctrl-btn small queue-menu-trigger {queueCount > 0 ? 'has-queue' : ''}"
    title={queueCount > 0 ? `播放列表（${queueCount} 首）` : '播放列表'}
    onclick={(e) => e.stopPropagation()}
  >
    <ListMusic size={16} />
  </Menu.Trigger>

  <Portal>
    <Menu.Positioner>
      <Menu.Content class="queue-menu-content card {immersive ? 'immersive' : ''}">
        <Menu.ItemGroup>
          <Menu.ItemGroupLabel class="queue-menu-header">
            <span class="queue-menu-title">
              播放列表
              <span class="queue-menu-count">{queueCount} 首</span>
            </span>
            {#if queueCount > 0}
              <button
                type="button"
                class="queue-menu-clear"
                onclick={(e) => {
                  e.stopPropagation();
                  clearQueue();
                }}
              >
                清空
              </button>
            {/if}
          </Menu.ItemGroupLabel>

          {#if queueCount === 0}
            <p class="queue-empty">当前没有播放列表，在搜索或发现页播放歌曲后会自动加入。</p>
          {:else}
            {#each player.queue as track, index (trackKey(track, index))}
              {@const active = isCurrentTrack(track)}
              {@const coverUrl = track.coverUrl?.trim() ?? ''}
              <Menu.Item
                value={`play:${index}`}
                closeOnSelect={false}
                class="queue-menu-item {active ? 'active' : ''}"
              >
                <span class="queue-index" aria-hidden="true">
                  {#if active && player.isPlaying}
                    <span class="playing-indicator">
                      <span></span><span></span><span></span>
                    </span>
                  {:else}
                    {index + 1}
                  {/if}
                </span>
                <span class="queue-cover" aria-hidden="true">
                  {#if coverUrl && !brokenCovers[coverUrl]}
                    <img
                      src={coverUrl}
                      alt=""
                      loading="lazy"
                      onerror={() => handleCoverError(coverUrl)}
                    />
                  {:else}
                    <Music size={14} />
                  {/if}
                </span>
                <span class="queue-meta">
                  <Menu.ItemText class="queue-song-title">{track.title}</Menu.ItemText>
                  <span class="queue-song-artist">{track.artist}</span>
                </span>
                <span class="queue-duration">{track.duration}</span>
                {#if active}
                  <span class="queue-playing-badge" aria-hidden="true">
                    <Play size={12} />
                  </span>
                {/if}
                <button
                  type="button"
                  class="queue-remove-btn"
                  aria-label={`从列表移除 ${track.title}`}
                  title="移除"
                  onclick={(e) => {
                    e.stopPropagation();
                    removeQueueTrack(index);
                  }}
                >
                  <Trash2 size={14} />
                </button>
              </Menu.Item>
            {/each}
          {/if}
        </Menu.ItemGroup>
      </Menu.Content>
    </Menu.Positioner>
  </Portal>
</Menu>

<style>
  :global(.queue-menu-content) {
    width: min(400px, calc(100vw - 280px));
    max-height: min(420px, calc(100vh - 160px));
    padding: 0;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    z-index: 200;
  }

  :global(.queue-menu-content.immersive) {
    background: rgba(28, 22, 20, 0.96);
    border-color: rgba(255, 255, 255, 0.1);
    color: rgba(255, 255, 255, 0.92);
  }

  :global(.queue-menu-header) {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 12px 14px 8px;
    margin: 0;
    border-bottom: 1px solid #f0f0f0;
  }

  :global(.queue-menu-content.immersive .queue-menu-header) {
    border-bottom-color: rgba(255, 255, 255, 0.08);
  }

  :global(.queue-menu-title) {
    display: flex;
    align-items: baseline;
    gap: 8px;
    font-size: 14px;
    font-weight: 600;
  }

  :global(.queue-menu-count) {
    font-size: 12px;
    font-weight: 400;
    color: #999;
  }

  :global(.queue-menu-content.immersive .queue-menu-count) {
    color: rgba(255, 255, 255, 0.45);
  }

  :global(.queue-menu-clear) {
    border: none;
    background: transparent;
    cursor: pointer;
    color: #888;
    border-radius: 6px;
    padding: 4px 8px;
    font-size: 12px;
    transition:
      color 0.15s ease,
      background 0.15s ease;
  }

  :global(.queue-menu-clear:hover) {
    color: #333;
    background: rgba(0, 0, 0, 0.06);
  }

  :global(.queue-menu-content.immersive .queue-menu-clear) {
    color: rgba(255, 255, 255, 0.55);
  }

  :global(.queue-menu-content.immersive .queue-menu-clear:hover) {
    color: #fff;
    background: rgba(255, 255, 255, 0.1);
  }

  :global(.queue-empty) {
    margin: 0;
    padding: 20px 16px 24px;
    font-size: 13px;
    line-height: 1.6;
    color: #999;
    text-align: center;
  }

  :global(.queue-menu-content.immersive .queue-empty) {
    color: rgba(255, 255, 255, 0.45);
  }

  :global(.queue-menu-content .queue-menu-item) {
    display: grid;
    grid-template-columns: 2rem 2.25rem minmax(0, 1fr) auto auto 2rem;
    align-items: center;
    gap: 8px;
    padding: 8px 10px;
    cursor: pointer;
    border-radius: 8px;
  }

  :global(.queue-menu-content .queue-menu-item.active) {
    background: #f0f4ff;
  }

  :global(.queue-menu-content.immersive .queue-menu-item.active) {
    background: rgba(102, 126, 234, 0.22);
  }

  :global(.queue-index) {
    font-size: 12px;
    color: #aaa;
    text-align: center;
    font-variant-numeric: tabular-nums;
  }

  :global(.queue-menu-content .queue-menu-item.active .queue-index) {
    color: #667eea;
  }

  :global(.queue-cover) {
    width: 36px;
    height: 36px;
    border-radius: 6px;
    background: #f0f0f0;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    color: #bbb;
    flex-shrink: 0;
  }

  :global(.queue-cover img) {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  :global(.queue-meta) {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  :global(.queue-menu-content .queue-song-title) {
    font-size: 13px;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  :global(.queue-song-artist) {
    font-size: 11px;
    color: #999;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  :global(.queue-menu-content.immersive .queue-song-artist) {
    color: rgba(255, 255, 255, 0.45);
  }

  :global(.queue-menu-content .queue-menu-item.active .queue-song-title),
  :global(.queue-menu-content .queue-menu-item.active .queue-song-artist),
  :global(.queue-menu-content .queue-menu-item.active .queue-duration) {
    color: #5b6ee8;
  }

  :global(.queue-menu-content.immersive .queue-menu-item.active .queue-song-title),
  :global(.queue-menu-content.immersive .queue-menu-item.active .queue-song-artist),
  :global(.queue-menu-content.immersive .queue-menu-item.active .queue-duration) {
    color: #a8b4ff;
  }

  :global(.queue-duration) {
    font-size: 11px;
    color: #999;
    font-variant-numeric: tabular-nums;
  }

  :global(.queue-playing-badge) {
    color: #667eea;
    display: flex;
    align-items: center;
  }

  :global(.queue-remove-btn) {
    width: 28px;
    height: 28px;
    border: none;
    background: transparent;
    color: #ccc;
    border-radius: 6px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    opacity: 0;
    transition:
      opacity 0.15s ease,
      color 0.15s ease,
      background 0.15s ease;
  }

  :global(.queue-menu-content .queue-menu-item:hover .queue-remove-btn),
  :global(.queue-remove-btn:focus-visible) {
    opacity: 1;
  }

  :global(.queue-remove-btn:hover) {
    color: #e74c3c;
    background: rgba(231, 76, 60, 0.08);
  }

  :global(.queue-menu-trigger[data-state='open']),
  :global(.queue-menu-trigger.has-queue) {
    color: #667eea;
  }

  :global(.playing-indicator) {
    display: inline-flex;
    align-items: flex-end;
    justify-content: center;
    gap: 2px;
    height: 12px;
    width: 100%;
  }

  :global(.playing-indicator span) {
    width: 2px;
    background: #667eea;
    border-radius: 1px;
    animation: bounce 0.8s ease-in-out infinite;
  }

  :global(.playing-indicator span:nth-child(1)) {
    height: 60%;
  }

  :global(.playing-indicator span:nth-child(2)) {
    height: 100%;
    animation-delay: 0.2s;
  }

  :global(.playing-indicator span:nth-child(3)) {
    height: 40%;
    animation-delay: 0.4s;
  }

  @keyframes bounce {
    0%,
    100% {
      transform: scaleY(1);
    }
    50% {
      transform: scaleY(0.4);
    }
  }
</style>
