<script lang="ts">
  import { AlertCircle, Info, CheckCircle2, X } from '@lucide/svelte';
  import { dismiss, toasts } from '@/stores/ui/toast';
</script>

<div class="toast-viewport" aria-live="polite" aria-atomic="true">
  {#each $toasts as toast (toast.id)}
    <div class="toast-item" class:error={toast.type === 'error'} class:success={toast.type === 'success'}>
      <span class="icon" aria-hidden="true">
        {#if toast.type === 'error'}
          <AlertCircle size={16} />
        {:else if toast.type === 'success'}
          <CheckCircle2 size={16} />
        {:else}
          <Info size={16} />
        {/if}
      </span>
      <span class="message">{toast.message}</span>
      <button type="button" class="close" onclick={() => dismiss(toast.id)} aria-label="关闭提示">
        <X size={14} />
      </button>
    </div>
  {/each}
</div>

<style>
  .toast-viewport {
    position: fixed;
    top: 16px;
    right: 16px;
    z-index: 200;
    display: flex;
    flex-direction: column;
    gap: 8px;
    pointer-events: none;
  }

  .toast-item {
    min-width: 220px;
    max-width: 360px;
    border-radius: 10px;
    padding: 10px 10px 10px 12px;
    background: #1f2937;
    color: #fff;
    box-shadow: 0 8px 20px rgba(0, 0, 0, 0.22);
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: center;
    gap: 8px;
    pointer-events: auto;
  }

  .toast-item.error {
    background: #b91c1c;
  }

  .toast-item.success {
    background: #166534;
  }

  .icon {
    display: inline-flex;
  }

  .message {
    font-size: 13px;
    line-height: 1.4;
  }

  .close {
    border: none;
    background: transparent;
    color: inherit;
    cursor: pointer;
    padding: 2px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }
</style>
