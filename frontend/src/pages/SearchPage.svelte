<script lang="ts">
  import { Button, Heading, Spinner } from 'flowbite-svelte';
  import { Music, Search } from '@lucide/svelte';
  import { router } from 'svelte-spa-router';
  import TrackList, { type TrackItem } from '@/components/TrackList.svelte';
  import { Search as searchApi, ListSources } from '../../wailsjs/go/main/App';
  import { music } from '../../wailsjs/go/models';
  import { buildSearchHref, parseSearchParams } from '@/lib/searchParams';

  type SongItem = music.SongItem;
  type SearchResult = music.SearchResult;

  const PLATFORM_LABELS: Record<string, string> = {
    wy: '网易云音乐',
    kw: '酷我音乐',
    kg: '酷狗音乐',
    tx: 'QQ音乐',
    mg: '咪咕音乐',
  };

  let songs = $state<SongItem[]>([]);
  let total = $state(0);
  let page = $state(1);
  let limit = $state(20);
  let sourcePlatform = $state('');
  let loading = $state(false);
  let error = $state('');
  let keyword = $state('');
  let currentSongId = $state<string | null>(null);
  let isPlaying = $state(false);
  let brokenCovers = $state<Record<string, true>>({});

  const hasKeyword = $derived(keyword.length > 0);
  const totalPages = $derived(Math.max(1, Math.ceil(total / limit)));
  const canPrev = $derived(page > 1);
  const canNext = $derived(page < totalPages);
  const platformLabel = $derived(PLATFORM_LABELS[sourcePlatform] ?? sourcePlatform);
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

  async function resolveSourceId(): Promise<string> {
    const sources = await ListSources();
    const ready = sources.find((item) => item.enabled && item.status === 'ready');
    if (!ready) {
      throw new Error('请先在设置中导入并启用音源');
    }
    return ready.id;
  }

  async function runSearch(q: string, pageNum: number) {
    keyword = q;
    page = pageNum;

    if (!q) {
      songs = [];
      total = 0;
      sourcePlatform = '';
      error = '';
      loading = false;
      return;
    }

    loading = true;
    error = '';
    try {
      const sourceId = await resolveSourceId();
      const result: SearchResult = await searchApi(sourceId, '', q, pageNum);
      songs = result.list ?? [];
      total = result.total ?? 0;
      page = result.page ?? pageNum;
      limit = result.limit ?? 20;
      sourcePlatform = result.source ?? '';
      brokenCovers = {};
    } catch (err) {
      songs = [];
      total = 0;
      error = errorMessage(err);
    } finally {
      loading = false;
    }
  }

  function markCoverBroken(url: string) {
    if (!url || brokenCovers[url]) {
      return;
    }
    brokenCovers = { ...brokenCovers, [url]: true };
  }

  function playTrack(track: TrackItem) {
    const id = String(track.id);
    if (currentSongId === id) {
      isPlaying = !isPlaying;
      return;
    }
    currentSongId = id;
    isPlaying = true;
  }

  function goToPage(nextPage: number) {
    if (!keyword || nextPage < 1 || nextPage > totalPages) {
      return;
    }
    window.location.hash = buildSearchHref(keyword, nextPage);
  }

  $effect(() => {
    if (router.location !== '/search') {
      return;
    }
    const { q, page: pageNum } = parseSearchParams(router.querystring);
    void runSearch(q, pageNum);
  });
</script>

<div class="search-page">
  <div class="section-header">
    <Heading tag="h2" class="text-2xl font-bold text-gray-800">
      <span class="title-row">
        <Search size={24} />
        搜索
      </span>
    </Heading>
    {#if hasKeyword && !loading && !error}
      <p class="result-meta">
        {#if platformLabel}
          <span class="platform-tag">{platformLabel}</span>
        {/if}
        共 {total} 条结果
      </p>
    {/if}
  </div>

  {#if !hasKeyword}
    <div class="empty-state">
      <Search size={48} class="text-gray-300" />
      <p>在顶部搜索框输入关键词后按 Enter</p>
    </div>
  {:else if loading}
    <div class="loading-state">
      <Spinner size="8" />
      <p>正在搜索「{keyword}」…</p>
    </div>
  {:else if error}
    <div class="error-state">
      <p>{error}</p>
    </div>
  {:else if songs.length === 0}
    <div class="empty-state">
      <Music size={48} class="text-gray-300" />
      <p>未找到与「{keyword}」相关的歌曲</p>
    </div>
  {:else}
    <TrackList
      {tracks}
      activeId={currentSongId}
      {isPlaying}
      {indexOffset}
      {brokenCovers}
      onSelect={playTrack}
      onCoverError={markCoverBroken}
      ariaLabel="搜索结果"
    />

    {#if totalPages > 1}
      <div class="pagination">
        <Button color="alternative" disabled={!canPrev} onclick={() => goToPage(page - 1)}>
          上一页
        </Button>
        <span class="page-info">第 {page} / {totalPages} 页</span>
        <Button color="alternative" disabled={!canNext} onclick={() => goToPage(page + 1)}>
          下一页
        </Button>
      </div>
    {/if}
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

  .title-row {
    display: inline-flex;
    align-items: center;
    gap: 8px;
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
  }
</style>
