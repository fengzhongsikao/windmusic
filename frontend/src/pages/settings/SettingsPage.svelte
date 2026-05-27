<!--
  设置页：洛雪自定义音源的导入、启用/禁用、删除及数据目录展示。
-->
<script lang="ts">
  import { Button, Toggle, Badge } from 'flowbite-svelte';
  import { FolderOpen, Plus, Settings, Trash2 } from '@lucide/svelte';
  import {
    DeleteSource,
    DisableSource,
    EnableSource,
    GetSourceDataDir,
    ImportSource,
    ListSources,
  } from '../../../wailsjs/go/main/App';
  import { music } from '../../../wailsjs/go/models';

  type SourceInfo = music.SourceInfo;
  type PlatformInfo = music.PlatformInfo;
  type BadgeColor = 'green' | 'yellow' | 'red';

  const STATUS_COLORS: Record<string, BadgeColor> = {
    ready: 'green',
    error: 'red',
  };

  let sources = $state<SourceInfo[]>([]);
  let dataDir = $state('');
  let loading = $state(false);
  let message = $state('');
  let error = $state('');

  const isEmpty = $derived(!loading && sources.length === 0);
  const hasSources = $derived(!loading && sources.length > 0);

  function errorMessage(err: unknown): string {
    return err instanceof Error ? err.message : String(err);
  }

  function statusColor(status: string): BadgeColor {
    return STATUS_COLORS[status] ?? 'yellow';
  }

  function formatPlatforms(source: SourceInfo): string {
    if (!source.platforms?.length) {
      return '无';
    }
    return source.platforms.map((item: PlatformInfo) => item.name || item.key).join('、');
  }

  function toggleChecked(event: Event): boolean {
    return (event.currentTarget as HTMLInputElement).checked;
  }

  async function refreshSources() {
    loading = true;
    error = '';
    try {
      sources = await ListSources();
      dataDir = await GetSourceDataDir();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      loading = false;
    }
  }

  async function handleImport() {
    message = '';
    error = '';
    try {
      const imported = await ImportSource();
      message = `已导入音源：${imported.name}`;
      await refreshSources();
    } catch (err) {
      const text = errorMessage(err);
      if (text !== 'import cancelled') {
        error = text;
      }
    }
  }

  async function handleToggle(source: SourceInfo, enabled: boolean) {
    error = '';
    try {
      if (enabled) {
        await EnableSource(source.id);
      } else {
        await DisableSource(source.id);
      }
      await refreshSources();
    } catch (err) {
      error = errorMessage(err);
      await refreshSources();
    }
  }

  async function handleDelete(source: SourceInfo) {
    if (!confirm(`确定删除音源「${source.name}」吗？`)) {
      return;
    }
    error = '';
    try {
      await DeleteSource(source.id);
      message = `已删除音源：${source.name}`;
      await refreshSources();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  $effect(() => {
    void refreshSources();
  });
</script>

<div class="settings-page">
  <div class="page-header">
    <h2 class="section-title">
      <Settings size={28} />
      设置
    </h2>
    <Button color="purple" onclick={handleImport}>
      <Plus size={16} class="mr-2" />
      导入音源
    </Button>
  </div>

  <section class="panel">
    <div class="panel-title">
      <FolderOpen size={18} />
      音源管理
    </div>
    <p class="panel-desc">
      支持导入 SixYin、Huibq、Flower 等主流 LX Music `.js` 音源。音源文件与配置会保存在本地目录，重启后自动加载。
    </p>
    {#if dataDir}
      <p class="data-dir">数据目录：{dataDir}</p>
    {/if}
    {#if message}
      <p class="feedback success">{message}</p>
    {/if}
    {#if error}
      <p class="feedback error">{error}</p>
    {/if}

    {#if loading}
      <p class="empty-state">加载中...</p>
    {:else if isEmpty}
      <p class="empty-state">还没有导入音源，点击右上角「导入音源」开始。</p>
    {:else if hasSources}
      <div class="source-list">
        {#each sources as source (source.id)}
          {@render sourceCard(source)}
        {/each}
      </div>
    {/if}
  </section>
</div>

{#snippet sourceCard(source: SourceInfo)}
  <article class="source-card">
    <div class="source-main">
      <div class="source-title-row">
        <h3>{source.name}</h3>
        <Badge color={statusColor(source.status)}>{source.status}</Badge>
      </div>
      {#if source.description}
        <p class="source-desc">{source.description}</p>
      {/if}
      <div class="source-meta">
        <span>版本：{source.version || '未知'}</span>
        <span>作者：{source.author || '未知'}</span>
        <span>平台：{formatPlatforms(source)}</span>
      </div>
      {#if source.error}
        <p class="source-error">{source.error}</p>
      {/if}
    </div>
    <div class="source-actions">
      <Toggle
        color="green"
        checked={source.enabled}
        onchange={(event) => handleToggle(source, toggleChecked(event))}
      >
        <span class="toggle-label" class:enabled={source.enabled}>
          {source.enabled ? '已启用' : '已禁用'}
        </span>
      </Toggle>
      <button
        type="button"
        class="icon-btn danger"
        aria-label="删除音源"
        title="删除音源"
        onclick={() => handleDelete(source)}
      >
        <Trash2 size={16} />
      </button>
    </div>
  </article>
{/snippet}

<style>
  .settings-page {
    max-width: 860px;
  }

  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
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
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0 0 8px;
    font-size: 16px;
    font-weight: 600;
    color: #333;
  }

  .panel-desc,
  .data-dir,
  .empty-state {
    margin: 0 0 12px;
    font-size: 14px;
    color: #666;
    line-height: 1.6;
  }

  .data-dir {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 12px;
    color: #888;
    word-break: break-all;
  }

  .feedback {
    margin: 0 0 12px;
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

  .source-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .source-card {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    padding: 16px;
    border-radius: 10px;
    background: #fff;
    border: 1px solid #ececec;
  }

  .source-main {
    min-width: 0;
    flex: 1;
  }

  .source-title-row {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 6px;
  }

  .source-title-row h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
    color: #333;
  }

  .source-desc {
    margin: 0 0 8px;
    font-size: 13px;
    color: #666;
  }

  .source-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    font-size: 12px;
    color: #888;
  }

  .source-error {
    margin: 8px 0 0;
    font-size: 12px;
    color: #b91c1c;
  }

  .source-actions {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-shrink: 0;
  }

  .toggle-label {
    font-size: 13px;
    color: #888;
  }

  .toggle-label.enabled {
    color: #15803d;
    font-weight: 500;
  }

  .icon-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    border: none;
    border-radius: 8px;
    background: transparent;
    color: #666;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .icon-btn:hover {
    background: rgba(0, 0, 0, 0.05);
  }

  .icon-btn.danger:hover {
    color: #dc2626;
    background: rgba(220, 38, 38, 0.08);
  }
</style>
