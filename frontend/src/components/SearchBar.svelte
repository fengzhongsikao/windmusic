<script lang="ts">
  import { Button, Input } from 'flowbite-svelte';
  import { Search, Settings, X } from '@lucide/svelte';
  import { link, push, router } from 'svelte-spa-router';
  import { buildSearchHref, parseSearchParams } from '@/lib/searchParams';

  let searchQuery = $state('');

  function clearSearch() {
    searchQuery = '';
    if (router.location === '/search') {
      void push('/search');
    }
  }

  function submitSearch() {
    const keyword = searchQuery.trim();
    if (!keyword) {
      return;
    }
    void push(buildSearchHref(keyword));
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
    if (q !== searchQuery) {
      searchQuery = q;
    }
  });
</script>

<div class="search-bar-row">
  <form class="search-form" onsubmit={(e) => { e.preventDefault(); submitSearch(); }}>
    <div class="search-input-wrap">
      <Input
        type="text"
        bind:value={searchQuery}
        placeholder="搜索音乐、歌手、专辑..."
        class="w-full ps-9"
        onkeydown={handleKeydown}
      >
        {#snippet left()}
          <Search size={18} class="text-gray-400" />
        {/snippet}
        {#snippet right()}
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
        {/snippet}
      </Input>
    </div>
    <Button type='submit' color='blue'>搜索</Button>
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
