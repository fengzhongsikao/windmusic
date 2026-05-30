<script lang="ts">
  import { Play } from '@lucide/svelte';

  interface Props {
    disabled?: boolean;
    onclick?: () => void;
  }

  let { disabled = false, onclick }: Props = $props();
</script>

<button type="button" class="play-all-btn" {disabled} {onclick}>
  <Play size={16} />
  <span>全部播放</span>
</button>

<style>
  .play-all-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 14px;
    border: none;
    border-radius: 999px;
    background: #667eea;
    color: #fff;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition:
      background 0.15s ease,
      opacity 0.15s ease;
  }

  .play-all-btn:hover:not(:disabled) {
    background: #5a6fd6;
  }

  .play-all-btn:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
</style>
