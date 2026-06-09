<script lang="ts">
  import { Plus, Settings, Trash2 } from '@lucide/svelte';
  import {
    addMetingURL,
    metingSettings,
    removeMetingURL,
    setActiveMetingURL,
  } from '@/stores/sources/meting.svelte';
  import { ClearLocalLibraryCache, GetSourceDataDir } from '../../../wailsjs/go/main/App';
  let dataDir = $state('');
  let message = $state('');
  let error = $state('');
  let draftURL = $state('');
  let saving = $state(false);
  let showClearCacheDialog = $state(false);
  let clearingCache = $state(false);

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

  async function confirmClearLocalCache() {
    if (clearingCache) return;
    clearingCache = true;
    try {
      await ClearLocalLibraryCache();
      showClearCacheDialog = false;
      error = '';
      message = '本地扫描缓存已清除，正在重新扫描';
    } catch (err) {
      error = errorMessage(err);
      message = '';
    } finally {
      clearingCache = false;
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

  <section class="panel">
    <h3 class="panel-title">本地音乐缓存</h3>
    <p class="hint">
      清除 <code>local-library.db</code> 中的扫描结果与封面缓存（<code>local-covers/</code>）。已添加的音乐文件夹不会受影响，清除后将自动重新扫描。
    </p>
    <button
      type="button"
      class="btn clear-cache-btn danger"
      disabled={saving || clearingCache}
      onclick={() => (showClearCacheDialog = true)}
    >
      <Trash2 size={16} />
      清除本地扫描缓存
    </button>
  </section>
</div>

{#if showClearCacheDialog}
  <div class="dialog-backdrop" role="presentation">
    <div class="alert-dialog" role="alertdialog" aria-modal="true" aria-labelledby="clear-cache-title">
      <h3 id="clear-cache-title">清除本地扫描缓存？</h3>
      <p>
        将清空 local-library.db 中的扫描记录、封面键与歌词缓存，并删除 local-covers 目录下的封面文件。此操作不可恢复，清除后会重新扫描已添加的文件夹。
      </p>
      <div class="dialog-actions">
        <button type="button" class="btn action-btn" onclick={() => (showClearCacheDialog = false)}>取消</button>
        <button
          type="button"
          class="btn action-btn danger"
          disabled={clearingCache}
          onclick={() => void confirmClearLocalCache()}
        >
          {clearingCache ? '清除中…' : '确定清除'}
        </button>
      </div>
    </div>
  </div>
{/if}

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

  .panel + .panel {
    margin-top: 16px;
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

  .hint code {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 12px;
    background: rgba(0, 0, 0, 0.05);
    padding: 1px 4px;
    border-radius: 4px;
  }

  .clear-cache-btn {
    padding: 8px 14px;
    background: rgba(239, 68, 68, 0.12);
    color: #dc2626;
  }

  .clear-cache-btn:hover:not(:disabled) {
    background: rgba(239, 68, 68, 0.2);
  }

  .clear-cache-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .dialog-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.35);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 50;
    padding: 16px;
  }

  .alert-dialog {
    width: min(420px, 100%);
    border-radius: 14px;
    background: #fff;
    padding: 18px;
    box-shadow: 0 12px 30px rgba(0, 0, 0, 0.16);
  }

  .alert-dialog h3 {
    margin: 0;
    font-size: 18px;
    color: #111827;
  }

  .alert-dialog p {
    margin: 8px 0 0;
    color: #4b5563;
    font-size: 14px;
    line-height: 1.5;
  }

  .dialog-actions {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }

  .dialog-actions .action-btn {
    padding: 8px 14px;
    background: #f3f4f6;
    color: #374151;
  }

  .dialog-actions .action-btn.danger {
    background: rgba(239, 68, 68, 0.12);
    color: #dc2626;
  }

  .dialog-actions .action-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
</style>
