<!--
  首页发现：按分类 Tab 展示推荐歌曲。
  数据走 App.Search（与搜索页同一后端），带 session 缓存与 Tab 预取。
-->
<script lang="ts">
  import { Button, ButtonGroup, Heading, Spinner } from 'flowbite-svelte';
  import TrackList from '@/components/TrackList.svelte';
  import type { TrackItem } from '@/lib/track';
  import { buildPlaybackContext, trackItemToPlayerTrack } from '@/lib/playerTrack';
  import { player, setQueue, togglePlayByTrack } from '@/stores/player.svelte';
  import { Search as searchApi, ListSources } from '../../../wailsjs/go/main/App';
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
    if (cachedSourceId) {
      return cachedSourceId;
    }
    const sources = await ListSources();
    const ready = sources.find((item) => item.enabled && item.status === 'ready');
    if (!ready) {
      throw new Error('请先在设置中导入并启用音源');
    }
    cachedSourceId = ready.id;
    return ready.id;
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
    const sourceId = cachedSourceId ?? '';
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
      // 复用搜索后端：platform 传空表示使用音源默认平台。
      const result = await searchApi(sourceId, '', keyword, 1);
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
    const sourceId = cachedSourceId ?? '';
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
    <Heading tag="h2" class="text-2xl font-bold text-gray-800">首页</Heading>
    <ButtonGroup>
      {#each tabs as tab}
        <Button
          color={activeTab === tab.id ? 'green' : 'alternative'}
          onclick={() => (activeTab = tab.id)}
        >
          {tab.label}
        </Button>
      {/each}
    </ButtonGroup>
  </div>

  {#if error}
    <div class="recommend-error">{error}</div>
  {:else if loading && tracks.length === 0}
    <div class="recommend-loading">
      <Spinner size="8" />
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
</style>
