<script lang="ts">
  import { Plus, Settings, Trash2 } from '@lucide/svelte';
  import {
    addMetingURL,
    metingSettings,
    removeMetingURL,
    setActiveMetingURL,
  } from '@/stores/sources/meting.svelte';
  import { GetSourceDataDir } from '../../../wailsjs/go/main/App';
  let dataDir = $state('');
  let message = $state('');
  let error = $state('');
  let draftURL = $state('');
  let saving = $state(false);

  const sources = $derived(metingSettings.urls);
  const activeURL = $derived(metingSettings.activeUrl);

  function errorMessage(err: unknown): string {
    return err instanceof Error ? err.message : String(err);
  }

  async function loadDataDir() {
    try {
      dataDir = await GetSourceDataDir();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  $effect(() => {
    void loadDataDir();
  });

  async function handleAdd() {
    saving = true;
    const addError = await addMetingURL(draftURL);
    saving = false;
    if (addError) {
      message = '';
      error = addError;
      return;
    }
    draftURL = '';
    error = '';
    message = '已添加';
  }

  async function handleRemove(url: string) {
    saving = true;
    try {
      await removeMetingURL(url);
      error = '';
      message = '已删除';
    } catch (err) {
      error = errorMessage(err);
      message = '';
    } finally {
      saving = false;
    }
  }

  async function handleSetActive(url: string) {
    saving = true;
    try {
      await setActiveMetingURL(url);
      error = '';
      message = '';
    } catch (err) {
      error = errorMessage(err);
    } finally {
      saving = false;
    }
  }
</script>

<div class="settings-page">
  <div class="page-header">
    <h2 class="section-title">
      <Settings size={28} />
      设置
    </h2>
  </div>

  <section class="panel">
    <h3 class="panel-title">Meting 源</h3>

    {#if sources.length > 0}
      <ul class="source-list">
        {#each sources as url (url)}
          <li class="source-item" class:active={url === activeURL}>
            <button
              type="button"
              class="source-url"
              title={url === activeURL ? '当前使用' : '设为当前'}
              disabled={saving}
              onclick={() => void handleSetActive(url)}
            >
              {url}
            </button>
            {#if url === activeURL}
              <span class="source-badge">当前</span>
            {/if}
            <button
              type="button"
              class="btn icon-btn danger"
              aria-label="删除"
              disabled={saving}
              onclick={() => void handleRemove(url)}
            >
              <Trash2 size={16} />
            </button>
          </li>
        {/each}
      </ul>
    {:else}
      <p class="hint">暂无 Meting 源</p>
    {/if}

    <div class="add-row">
      <input
        class="meting-input"
        type="url"
        placeholder="https://meting.mikus.ink"
        bind:value={draftURL}
        disabled={saving}
        onkeydown={(e) => {
          if (e.key === 'Enter') void handleAdd();
        }}
      />
      <button type="button" class="btn add-btn" disabled={saving} onclick={() => void handleAdd()}>
        <Plus size={16} />
        添加
      </button>
    </div>

    {#if dataDir}
      <p class="data-dir">数据目录：{dataDir}</p>
    {/if}
    {#if message}
      <p class="feedback success">{message}</p>
    {/if}
    {#if error}
      <p class="feedback error">{error}</p>
    {/if}
  </section>
</div>

<style>
  .settings-page {
    max-width: 860px;
  }

  .page-header {
    margin-bottom: 24px;
  }

  .section-title {
    display: flex;
    align-items: center;
    gap: 10px;
    margin: 0;
    font-size: 24px;
    font-weight: 700;
    color: #333;
  }

  .panel {
    border: 1px solid #eee;
    border-radius: 12px;
    padding: 20px;
    background: #fafafa;
  }

  .panel-title {
    margin: 0 0 16px;
    font-size: 16px;
    font-weight: 600;
    color: #333;
  }

  .hint {
    margin: 0 0 12px;
    font-size: 14px;
    color: #999;
  }

  .source-list {
    list-style: none;
    margin: 0 0 16px;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .source-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 10px;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    background: #fff;
  }

  .source-item.active {
    border-color: #667eea;
    background: #f8f9ff;
  }

  .source-url {
    flex: 1;
    min-width: 0;
    border: none;
    background: none;
    padding: 0;
    font: inherit;
    font-size: 13px;
    color: #374151;
    text-align: left;
    cursor: pointer;
    word-break: break-all;
  }

  .source-url:hover {
    color: #667eea;
  }

  .source-badge {
    flex-shrink: 0;
    font-size: 11px;
    color: #667eea;
    background: rgba(102, 126, 234, 0.12);
    padding: 2px 8px;
    border-radius: 999px;
  }

  .add-row {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 10px;
  }

  .meting-input {
    border: 1px solid #ddd;
    border-radius: 8px;
    padding: 8px 10px;
    font-size: 14px;
    background: #fff;
  }

  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    font-size: 13px;
  }

  .add-btn {
    padding: 8px 14px;
    background: #667eea;
    color: #fff;
  }

  .add-btn:hover {
    background: #5a6fd6;
  }

  .icon-btn {
    flex-shrink: 0;
    width: 32px;
    height: 32px;
    padding: 0;
    background: #f3f4f6;
    color: #6b7280;
  }

  .icon-btn.danger:hover {
    background: rgba(239, 68, 68, 0.12);
    color: #dc2626;
  }

  .data-dir {
    margin: 16px 0 0;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 12px;
    color: #888;
    word-break: break-all;
  }

  .feedback {
    margin: 12px 0 0;
    padding: 10px 12px;
    border-radius: 8px;
    font-size: 13px;
  }

  .feedback.success {
    background: rgba(34, 197, 94, 0.12);
    color: #15803d;
  }

  .feedback.error {
    background: rgba(239, 68, 68, 0.12);
    color: #b91c1c;
  }
</style>
