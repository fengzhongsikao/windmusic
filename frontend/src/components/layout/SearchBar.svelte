<script lang="ts">
  import { Search, Settings, X } from '@lucide/svelte';
  import { link, router } from 'svelte-spa-router';
  import { buildSearchHref, parseSearchParams } from '@/lib/sources/searchParams';

  let searchQuery = $state('');

  function clearSearch() {
    searchQuery = '';
    if (router.location === '/search') {
      window.location.hash = '/search';
    }
  }

  function submitSearch() {
    const keyword = searchQuery.trim();
    if (!keyword) {
      return;
    }
    window.location.hash = buildSearchHref(keyword);
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      event.preventDefault();
      submitSearch();
    }
  }

  $effect(() => {
    if (router.location !== '/search') {
      return;
    }
    const { q } = parseSearchParams(router.querystring);
    // 仅在路由参数变化时同步，避免输入过程中被旧 q 覆盖。
    searchQuery = q;
  });
</script>

<div class="search-bar-row">
  <form class="search-form" onsubmit={(e) => { e.preventDefault(); submitSearch(); }}>
    <div class="search-input-wrap">
      <div class="search-input">
        <Search size={18} style="position:absolute;left:10px;color:#9ca3af;pointer-events:none;" />
        <input
        type="text"
        bind:value={searchQuery}
        placeholder="搜索音乐、歌手、专辑..."
        class="search-input-field"
        onkeydown={handleKeydown}
      />
        {#if searchQuery}
          <button
            type="button"
            class="clear-btn"
            onclick={clearSearch}
            aria-label="清除搜索"
          >
            <X size={16} />
          </button>
        {/if}
      </div>
    </div>
    <button type="submit" class="btn preset-filled-success-500">搜索</button>
  </form>
  <a href="/settings" use:link class="settings-btn" aria-label="设置" title="设置">
    <Settings size={20} />
  </a>
</div>

<style>
  .search-bar-row {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .search-form {
    display: flex;
    align-items: center;
    gap: 8px;
    flex: 1;
    width: 100%;
    max-width: 560px;
  }

  .search-input-wrap {
    flex: 1;
    min-width: 0;
  }

  .search-input {
    position: relative;
    display: flex;
    align-items: center;
    width: 100%;
  }

  .search-input-field {
    width: 100%;
    height: 40px;
    border: 1px solid #d1d5db;
    border-radius: 10px;
    padding: 0 34px 0 34px;
    background: #fff;
    color: #111827;
    outline: none;
    transition: border-color 0.2s ease, box-shadow 0.2s ease;
  }

  .search-input-field:focus {
    border-color: #667eea;
    box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.15);
  }

  .clear-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: transparent;
    border-radius: 999px;
    padding: 4px;
    color: #666;
    cursor: pointer;
    transition: all 0.2s ease;
    position: absolute;
    right: 8px;
  }

  .clear-btn:hover {
    background: rgba(0, 0, 0, 0.06);
    color: #333;
  }

  .settings-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    width: 36px;
    height: 36px;
    border-radius: 8px;
    color: #666;
    text-decoration: none;
    transition: all 0.2s ease;
  }

  .settings-btn:hover {
    background: rgba(0, 0, 0, 0.04);
    color: #667eea;
  }

</style>
