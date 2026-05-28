<!--
  首页发现：按分类 Tab 展示推荐歌曲。
  数据走 App.Search（与搜索页同一后端），带 session 缓存与 Tab 预取。
-->
<script lang="ts">
  import TrackList from '@/components/TrackList.svelte';
  import type { TrackItem } from '@/lib/track';
  import { buildPlaybackContext, trackItemToPlayerTrack } from '@/lib/playerTrack';
  import { getMetingPlatform, getMetingURL, metingSourceId } from '@/lib/meting';
  import { player, setQueue, togglePlayByTrack } from '@/stores/player.svelte';
  import { Search as searchApi } from '../../../wailsjs/go/main/App';
  import { music } from '../../../wailsjs/go/models';

  let activeTab = $state('all');

  const tabs = [
    { id: 'all', label: '全部' },
    { id: 'chinese', label: '华语' },
    { id: 'pop', label: '流行' },
    { id: 'rock', label: '摇滚' },
    { id: 'electronic', label: '电子' },
  ];

  // 用和搜索页同一套后端 `App.Search` 拉取真实列表。
  // 这里的关键词会在后端做平台搜索，并返回结果列表。
  const CATEGORY_KEYWORDS: Record<string, string> = {
    all: '热门',
    chinese: '华语',
    pop: '流行',
    rock: '摇滚',
    electronic: '电子',
  };

  const RECOMMEND_LIMIT = 10;
  const CACHE_TTL_MS = 5 * 60 * 1000;
  const CACHE_PREFIX = 'discover-recommend:';

  let songs = $state<music.SongItem[]>([]);
  let loading = $state(false);
  let error = $state('');

  let currentSongId = $derived(player.currentSong.id);

  let cachedSourceId = $state<string | null>(null);
  let recommendRequestId = 0;
  const inFlightRequests = new Map<string, Promise<music.SongItem[]>>();
  let hasPrefetchedTabs = false;

  type RecommendCacheEntry = {
    songs: music.SongItem[];
    cachedAt: number;
  };

  async function resolveSourceId(): Promise<string> {
    const metingURL = getMetingURL();
    if (metingURL) {
      return metingSourceId(metingURL);
    }
    return 'builtin::network';
  }

  function resolvePlatform(): string {
    const metingURL = getMetingURL();
    if (metingURL) {
      return getMetingPlatform();
    }
    return '';
  }

  const tracks = $derived<TrackItem[]>(
    songs.slice(0, RECOMMEND_LIMIT).map((song) => ({
      id: song.id,
      title: song.name,
      artist: song.singer,
      album: song.album,
      duration: song.interval ?? '—',
      coverUrl: song.img?.trim() || undefined,
    })),
  );

  function resolvePlayerTrack(track: TrackItem) {
    const song = songs.find((item) => String(item.id) === String(track.id));
    const metingURL = getMetingURL();
    const sourceId = metingURL ? metingSourceId(metingURL) : '';
    return trackItemToPlayerTrack(track, buildPlaybackContext(song, sourceId, ''));
  }

  function playTrack(track: TrackItem) {
    const playerTrack = resolvePlayerTrack(track);
    console.info('[发现页] 点击歌曲，切换播放状态', {
      id: playerTrack.id,
      title: playerTrack.title,
      artist: playerTrack.artist,
      album: playerTrack.album,
      coverUrl: playerTrack.coverUrl,
      playback: playerTrack.playback,
    });
    togglePlayByTrack(playerTrack);
  }

  async function runRecommend(tabId: string) {
    const keyword = CATEGORY_KEYWORDS[tabId] ?? CATEGORY_KEYWORDS.all;
    const requestId = ++recommendRequestId;
    const cacheKey = `${CACHE_PREFIX}${tabId}`;

    if (typeof window !== 'undefined') {
      const raw = window.sessionStorage.getItem(cacheKey);
      if (raw) {
        try {
          const parsed = JSON.parse(raw) as RecommendCacheEntry;
          if (Array.isArray(parsed.songs) && Date.now()- parsed.cachedAt < CACHE_TTL_MS) {
            songs = parsed.songs;
            error = '';
            loading = false;
            return;
          }
        } catch {
          // ignore malformed cache and continue fetching
        }
      }
    }

    loading = true;
    error = '';

    try {
      const resultSongs = await getRecommendSongs(tabId, keyword);
      if (requestId !== recommendRequestId) return;
      songs = resultSongs;
    } catch (err) {
      if (requestId !== recommendRequestId) return;
      error = err instanceof Error ? err.message : String(err);
    } finally {
      if (requestId !== recommendRequestId) return;
      loading = false;
    }
  }

  async function getRecommendSongs(tabId: string, keyword: string): Promise<music.SongItem[]> {
    const cacheKey = `${CACHE_PREFIX}${tabId}`;
    const existing = inFlightRequests.get(cacheKey);
    if (existing) {
      return existing;
    }

    const request = (async () => {
      const sourceId = await resolveSourceId();
      const result = await searchApi(sourceId, resolvePlatform(), keyword, 1);
      const resultSongs = result.list ?? [];
      if (typeof window !== 'undefined') {
        const entry: RecommendCacheEntry = {
          songs: resultSongs,
          cachedAt: Date.now(),
        };
        window.sessionStorage.setItem(cacheKey, JSON.stringify(entry));
      }
      return resultSongs;
    })();

    inFlightRequests.set(cacheKey, request);
    try {
      return await request;
    } finally {
      inFlightRequests.delete(cacheKey);
    }
  }

  async function prefetchAllTabs() {
    if (hasPrefetchedTabs) {
      return;
    }
    hasPrefetchedTabs = true;

    for (const tab of tabs) {
      const cacheKey = `${CACHE_PREFIX}${tab.id}`;
      if (typeof window !== 'undefined') {
        const raw = window.sessionStorage.getItem(cacheKey);
        if (raw) {
          try {
            const parsed = JSON.parse(raw) as RecommendCacheEntry;
            if (Array.isArray(parsed.songs) && Date.now() - parsed.cachedAt < CACHE_TTL_MS) {
              continue;
            }
          } catch {
            // ignore malformed cache and fetch again
          }
        }
      }
      const keyword = CATEGORY_KEYWORDS[tab.id] ?? CATEGORY_KEYWORDS.all;
      void getRecommendSongs(tab.id, keyword);
    }
  }

  // 初次加载 + Tab 切换都会刷新推荐
  $effect(() => {
    void runRecommend(activeTab);
    void prefetchAllTabs();
  });

  $effect(() => {
    const metingURL = getMetingURL();
    const sourceId = metingURL ? metingSourceId(metingURL) : '';
    setQueue(
      songs.map((song) =>
        trackItemToPlayerTrack(
          {
            id: song.id,
            title: song.name,
            artist: song.singer,
            album: song.album,
            duration: song.interval ?? '—',
            coverUrl: song.img?.trim() || undefined,
          },
          buildPlaybackContext(song, sourceId, ''),
        ),
      ),
    );
  });
</script>

<div class="home-page">
  <div class="section-header">
    <div class="tab-group">
      {#each tabs as tab}
        <button
          type="button"
          class="tab-button"
          class:active={activeTab === tab.id}
          onclick={() => (activeTab = tab.id)}
        >
          {tab.label}
        </button>
      {/each}
    </div>
  </div>

  {#if error}
    <div class="recommend-error">{error}</div>
  {:else if loading && tracks.length === 0}
    <div class="recommend-loading">
      <span class="loading-spinner" aria-hidden="true"></span>
      <p>正在加载推荐…</p>
    </div>
  {:else}
    <TrackList {tracks} activeId={currentSongId} isPlaying={player.isPlaying} onSelect={playTrack} />
  {/if}
</div>

<style>
  .home-page {
    padding: 0;
  }

  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 24px;
    flex-wrap: wrap;
    gap: 16px;
  }

  .page-title {
    margin: 0;
    font-size: 1.5rem;
    line-height: 2rem;
    font-weight: 700;
    color: #1f2937;
  }

  .tab-group {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .tab-button {
    border: 1px solid #d1d5db;
    border-radius: 999px;
    background: #fff;
    color: #4b5563;
    padding: 6px 14px;
    font-size: 14px;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .tab-button:hover {
    border-color: #aeb5c1;
    color: #1f2937;
  }

  .tab-button.active {
    border-color: #22c55e;
    background: rgba(34, 197, 94, 0.12);
    color: #15803d;
  }

  .recommend-loading,
  .recommend-error {
    padding: 64px 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-direction: column;
    gap: 12px;
    color: #999;
    text-align: center;
  }

  .recommend-error {
    color: #dc2626;
  }

  .loading-spinner {
    width: 32px;
    height: 32px;
    border-radius: 999px;
    border: 3px solid rgba(0, 0, 0, 0.1);
    border-top-color: #667eea;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
