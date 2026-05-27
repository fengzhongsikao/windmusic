<script lang="ts">
  import { Input } from 'flowbite-svelte';
  import { Search, Settings, X } from '@lucide/svelte';
  import { link } from 'svelte-spa-router';

  let searchQuery = $state('');

  function clearSearch() {
    searchQuery = '';
  }
</script>

<div class="search-bar-row">
  <div class="search-bar-wrap">
    <Input
      type="search"
      bind:value={searchQuery}
      placeholder="搜索音乐、歌手、专辑..."
      class="w-full ps-9"
    >
      {#snippet left()}
        <Search size={18} class="text-gray-400" />
      {/snippet}
      {#snippet right()}
        {#if searchQuery}
          <button
            type="button"
            class="flex items-center justify-center rounded-full p-1 text-gray-500 hover:bg-gray-100 hover:text-gray-700"
            onclick={clearSearch}
            aria-label="清除搜索"
          >
            <X size={16} />
          </button>
        {/if}
      {/snippet}
    </Input>
  </div>
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

  .search-bar-wrap {
    flex: 1;
    width: 100%;
    max-width: 480px;
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
