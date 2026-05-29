<!--
  搜索页：根据 URL 查询参数 q、page 调用 App.Search，支持分页与结果缓存。
-->
<script lang="ts">
  import { Music, Search } from '@lucide/svelte';
  import { router } from 'svelte-spa-router';
  import TrackList from '@/components/TrackList.svelte';
  import type { TrackItem } from '@/lib/track';
  import { buildPlaybackContext, trackItemToPlayerTrack } from '@/lib/playerTrack';
  import {
    getMetingPlatform,
    getMetingURL,
    metingSourceId,
    type MetingPlatform,
  } from '@/lib/meting';
  import PlayAllButton from '@/components/PlayAllButton.svelte';
  import {
    player,
    playAllTracks,
    togglePlayByTrack,
  } from '@/stores/player.svelte';
  import { Search as searchApi } from '../../../wailsjs/go/main/App';
  import { music } from '../../../wailsjs/go/models';
  import { buildSearchHref, parseSearchParams } from '@/lib/searchParams';

  type SongItem = music.SongItem;
  type SearchResult = music.SearchResult;

  const PLATFORM_LABELS: Record<MetingPlatform, string> = {
    netease: '网易云',
  };

  let songs = $state<SongItem[]>([]);
  let total = $state(0);
  let page = $state(1);
  let limit = $state(20);
  let sourcePlatform = $state('');
  let loading = $state(false);
  let pageLoading = $state(false);
  let error = $state('');
  let keyword = $state('');
  let selectedPlatform = $state<MetingPlatform>('netease');
  let currentSongId = $derived(player.currentSong.id);
  let brokenCovers = $state<Record<string, true>>({});

  let cachedSourceId = $state<string | null>(null);
  let searchRequestId = 0;

  type PageCacheEntry = {
    songs: SongItem[];
    total: number;
    page: number;
    limit: number;
    sourcePlatform: string;
  };

  const PAGE_CACHE_MAX = 32;
  const pageCache = new Map<string, PageCacheEntry>();

  const hasKeyword = $derived(keyword.length > 0);
  const totalPages = $derived(Math.max(1, Math.ceil(total / limit)));
  const canPrev = $derived(page > 1);
  const canNext = $derived(page < totalPages);
  const platformLabel = $derived(
    sourcePlatform in PLATFORM_LABELS
      ? PLATFORM_LABELS[sourcePlatform as MetingPlatform]
      : sourcePlatform,
  );
  const indexOffset = $derived((page - 1) * limit);

  const tracks = $derived<TrackItem[]>(
    songs.map((song) => ({
      id: song.id,
      title: song.name,
      artist: song.singer,
      album: song.album,
      duration: song.interval ?? '—',
      coverUrl: song.img?.trim() || undefined,
    })),
  );

  function errorMessage(err: unknown): string {
    return err instanceof Error ? err.message : String(err);
  }

  function pageCacheKey(sourceId: string, platform: string, q: string, pageNum: number) {
    return `${sourceId}|${platform}|${q}|${pageNum}`;
  }

  function readPageCache(key: string): PageCacheEntry | undefined {
    const entry = pageCache.get(key);
    if (!entry) {
      return undefined;
    }
    pageCache.delete(key);
    pageCache.set(key, entry);
    return entry;
  }

  function writePageCache(key: string, entry: PageCacheEntry) {
    pageCache.delete(key);
    pageCache.set(key, entry);
    while (pageCache.size > PAGE_CACHE_MAX) {
      const oldest = pageCache.keys().next().value;
      if (oldest === undefined) {
        break;
      }
      pageCache.delete(oldest);
    }
  }

  function clearPageCache() {
    pageCache.clear();
  }

  function applySearchResult(result: SearchResult, pageNum: number) {
    songs = result.list ?? [];
    total = result.total ?? 0;
    page = result.page ?? pageNum;
    limit = result.limit ?? 20;
    sourcePlatform = result.source ?? '';
    brokenCovers = {};
  }

  async function resolveSourceId(): Promise<string> {
    const metingURL = getMetingURL();
    if (metingURL) {
      return metingSourceId(metingURL);
    }
    return 'builtin::network';
  }

  function resolvePlatform(): string {
    return selectedPlatform;
  }

  async function runSearch(q: string, pageNum: number) {
    const prevKeyword = keyword;
    const isSameQuery = q === prevKeyword && q !== '';
    const hasExistingResults = isSameQuery && songs.length > 0;

    if (!q) {
      error = '';
      loading = false;
      pageLoading = false;
      return;
    }

    keyword = q;
    page = pageNum;

    if (!isSameQuery) {
      clearPageCache();
    }

    const requestId = ++searchRequestId;

    let sourceId: string;
    try {
      sourceId = await resolveSourceId();
    } catch (err) {
      if (requestId !== searchRequestId) {
        return;
      }
      songs = [];
      total = 0;
      error = errorMessage(err);
      loading = false;
      pageLoading = false;
      return;
    }

    const platform = isSameQuery && sourcePlatform ? sourcePlatform : resolvePlatform();
    const cacheKey = pageCacheKey(sourceId, platform, q, pageNum);
    const cached = readPageCache(cacheKey);
    if (cached) {
      songs = cached.songs;
      total = cached.total;
      page = cached.page;
      limit = cached.limit;
      sourcePlatform = cached.sourcePlatform;
      error = '';
      loading = false;
      pageLoading = false;
      return;
    }

    if (hasExistingResults) {
      pageLoading = true;
    } else {
      loading = true;
    }
    error = '';

    try {
      const result: SearchResult = await searchApi(sourceId, platform, q, pageNum);
      if (requestId !== searchRequestId) {
        return;
      }
      applySearchResult(result, pageNum);
      writePageCache(cacheKey, {
        songs,
        total,
        page,
        limit,
        sourcePlatform,
      });
    } catch (err) {
      if (requestId !== searchRequestId) {
        return;
      }
      if (!hasExistingResults) {
        songs = [];
        total = 0;
      }
      error = errorMessage(err);
    } finally {
      if (requestId === searchRequestId) {
        loading = false;
        pageLoading = false;
      }
    }
  }

  function markCoverBroken(url: string) {
    if (!url || brokenCovers[url]) {
      return;
    }
    brokenCovers = { ...brokenCovers, [url]: true };
  }

  function resolvePlayerTrack(track: TrackItem) {
    const song = songs.find((item) => String(item.id) === String(track.id));
    const metingURL = getMetingURL();
    const sourceId = metingURL ? metingSourceId(metingURL) : '';
    return trackItemToPlayerTrack(
      track,
      buildPlaybackContext(song, sourceId, sourcePlatform),
    );
  }

  function allPlayerTracks(): ReturnType<typeof trackItemToPlayerTrack>[] {
    const metingURL = getMetingURL();
    const sourceId = metingURL ? metingSourceId(metingURL) : '';
    return songs.map((song) =>
      trackItemToPlayerTrack(
        {
          id: song.id,
          title: song.name,
          artist: song.singer,
          album: song.album,
          duration: song.interval ?? '—',
          coverUrl: song.img?.trim() || undefined,
        },
        buildPlaybackContext(song, sourceId, sourcePlatform),
      ),
    );
  }

  function playTrack(track: TrackItem) {
    togglePlayByTrack(resolvePlayerTrack(track));
  }

  function playAll() {
    playAllTracks(allPlayerTracks());
  }

  function goToPage(nextPage: number) {
    if (!keyword || pageLoading || nextPage < 1 || nextPage > totalPages) {
      return;
    }
    window.location.hash = buildSearchHref(keyword, nextPage);
  }

  $effect(() => {
    if (router.location !== '/search') {
      return;
    }
    selectedPlatform = getMetingPlatform();
    const { q, page: pageNum } = parseSearchParams(router.querystring);
    void runSearch(q, pageNum);
  });

</script>

<div class="search-page">
  <div class="section-header">
    <div class="platform-picker">
      <span class="platform-label">平台</span>
      <span class="platform-tag">{PLATFORM_LABELS.netease}</span>
    </div>
    {#if hasKeyword && !loading && !error && songs.length > 0}
      <div class="header-actions">
        <p class="result-meta">
          {#if platformLabel}
            <span class="platform-tag">{platformLabel}</span>
          {/if}
          共 {total} 条结果
        </p>
        <PlayAllButton onclick={playAll} />
      </div>
    {/if}
  </div>

  {#if !hasKeyword}
    <div class="empty-state">
      <Search size={48} class="text-gray-300" />
      <p>在顶部搜索框输入关键词后按 Enter</p>
    </div>
  {:else if loading && songs.length === 0}
    <div class="loading-state">
      <span class="loading-spinner" aria-hidden="true"></span>
      <p>正在搜索「{keyword}」…</p>
    </div>
  {:else if error && songs.length === 0}
    <div class="error-state">
      <p>{error}</p>
    </div>
  {:else if songs.length === 0}
    <div class="empty-state">
      <Music size={48} class="text-gray-300" />
      <p>未找到与「{keyword}」相关的歌曲</p>
    </div>
  {:else}
    <div class="results-panel" class:page-loading={pageLoading}>
      {#if pageLoading}
        <div class="page-loading-bar" aria-hidden="true"></div>
      {/if}
      {#if error}
        <p class="inline-error">{error}</p>
      {/if}
      <TrackList
        {tracks}
        activeId={currentSongId}
        isPlaying={player.isPlaying}
        {indexOffset}
        {brokenCovers}
        onSelect={playTrack}
        onCoverError={markCoverBroken}
        ariaLabel="搜索结果"
      />

      {#if totalPages > 1}
        <div class="pagination">
          <button
            type="button"
            class="btn pagination-button"
            disabled={!canPrev || pageLoading}
            onclick={() => goToPage(page - 1)}
          >
            上一页
          </button>
          <span class="page-info">
            {#if pageLoading}
              加载中…
            {:else}
              第 {page} / {totalPages} 页
            {/if}
          </span>
          <button
            type="button"
            class="btn pagination-button"
            disabled={!canNext || pageLoading}
            onclick={() => goToPage(page + 1)}
          >
            下一页
          </button>
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .search-page {
    padding: 0;
  }

  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 24px;
    flex-wrap: wrap;
    gap: 12px;
  }

  .platform-picker {
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }

  .platform-label {
    font-size: 13px;
    color: #666;
  }

  .title-row {
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }

  .page-title {
    margin: 0;
    font-size: 1.5rem;
    line-height: 2rem;
    font-weight: 700;
    color: #1f2937;
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  .result-meta {
    margin: 0;
    font-size: 14px;
    color: #666;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .platform-tag {
    display: inline-block;
    padding: 2px 8px;
    border-radius: 999px;
    background: rgba(102, 126, 234, 0.1);
    color: #667eea;
    font-size: 12px;
  }

  .empty-state,
  .loading-state,
  .error-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 16px;
    padding: 64px 24px;
    color: #999;
    text-align: center;
  }

  .error-state {
    color: #dc2626;
  }

  .results-panel {
    position: relative;
  }

  .results-panel.page-loading {
    pointer-events: none;
    opacity: 0.72;
  }

  .page-loading-bar {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 2px;
    background: linear-gradient(90deg, transparent, #667eea, transparent);
    background-size: 200% 100%;
    animation: page-load 0.9s ease-in-out infinite;
    z-index: 1;
  }

  @keyframes page-load {
    0% {
      background-position: 200% 0;
    }
    100% {
      background-position: -200% 0;
    }
  }

  .inline-error {
    margin: 0 0 12px;
    font-size: 14px;
    color: #dc2626;
  }

  .pagination {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 16px;
    margin-top: 24px;
  }

  .page-info {
    font-size: 14px;
    color: #666;
    min-width: 7rem;
    text-align: center;
  }

  .loading-spinner {
    width: 32px;
    height: 32px;
    border-radius: 999px;
    border: 3px solid rgba(0, 0, 0, 0.1);
    border-top-color: #667eea;
    animation: spin 0.8s linear infinite;
  }

  .pagination-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
