<!-- 设置页：仅保留 Meting 网络源配置 -->
<script lang="ts">
  import { FolderOpen, Settings } from '@lucide/svelte';
  import { getMetingURL, setMetingURL } from '@/lib/meting';
  import { GetSourceDataDir } from '../../../wailsjs/go/main/App';

  let dataDir = $state('');
  let message = $state('');
  let error = $state('');
  let metingURL = $state('');

  function errorMessage(err: unknown): string {
    return err instanceof Error ? err.message : String(err);
  }

  async function loadDataDir() {
    error = '';
    try {
      dataDir = await GetSourceDataDir();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  $effect(() => {
    metingURL = getMetingURL();
    void loadDataDir();
  });

  function saveMetingSettings() {
    setMetingURL(metingURL);
    metingURL = getMetingURL();
    message = metingURL ? '已启用 Meting 源' : '已关闭 Meting 源';
    error = '';
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
    <div class="panel-title">
      <FolderOpen size={18} />
      Meting 源
    </div>
    <p class="panel-desc">
      填写 Meting 服务地址后，将只使用 Meting（网络源）进行搜索/播放/歌词。
    </p>
    <div class="meting-row">
      <input
        class="meting-input"
        type="text"
        placeholder="例如：https://meting.mikus.ink"
        bind:value={metingURL}
      />
      <button type="button" class="btn save-btn" onclick={saveMetingSettings}>保存</button>
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
    <p class="empty-state">已移除 JS 音源管理，仅保留 Meting 网络源。</p>
  </section>
</div>

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

  .save-btn {
    display: inline-flex;
    align-items: center;
  }

  .panel {
    border: 1px solid #eee;
    border-radius: 12px;
    padding: 20px;
    background: #fafafa;
  }

  .meting-row {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 10px;
    margin-bottom: 12px;
  }

  .meting-input {
    border: 1px solid #ddd;
    border-radius: 8px;
    padding: 8px 10px;
    font-size: 14px;
    background: #fff;
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

</style>
