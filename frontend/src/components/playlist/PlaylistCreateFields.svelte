<script lang="ts">
  import { createUserPlaylist, playlistCreateErrorMessage } from '@/lib/playlists';
  import { error as toastError } from '@/stores/ui/toast';

  interface Props {
    onCreated?: () => void;
  }

  let { onCreated }: Props = $props();

  let name = $state('');
  let saving = $state(false);
  let inputEl = $state<HTMLInputElement | null>(null);

  export function reset() {
    name = '';
    saving = false;
  }

  export function focusInput() {
    inputEl?.focus();
  }

  async function submit() {
    const trimmed = name.trim();
    if (!trimmed || saving) return;

    saving = true;
    try {
      await createUserPlaylist(trimmed);
      reset();
      onCreated?.();
    } catch (err) {
      toastError(playlistCreateErrorMessage(err));
    } finally {
      saving = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault();
      e.stopPropagation();
      void submit();
    }
  }
</script>

<div class="playlist-create-fields" onpointerdown={(e) => e.stopPropagation()}>
  <input
    bind:this={inputEl}
    class="input"
    type="text"
    maxlength={40}
    placeholder="输入歌单名称"
    bind:value={name}
    disabled={saving}
    onkeydown={handleKeydown}
  />
  <button
    type="button"
    class="btn preset-filled"
    disabled={saving || !name.trim()}
    onclick={() => void submit()}
  >
    {saving ? '创建中…' : '创建'}
  </button>
</div>

<style>
  .playlist-create-fields {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 8px 12px 12px;
  }

  .input {
    width: 100%;
    padding: 8px 10px;
    border: 1px solid #e5e5e5;
    border-radius: 8px;
    font-size: 13px;
    color: #333;
    background: #fafafa;
    outline: none;
  }

  .input:focus {
    border-color: #667eea;
    box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.12);
    background: #fff;
  }

  .input:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }
</style>
