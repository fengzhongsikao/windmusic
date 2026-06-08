<script lang="ts">
  import { Settings, ChevronDown, Check } from '@lucide/svelte';
  import { WAVEFORM_SPREAD_OPTIONS } from '@/lib/waveformSpread';
  import {
    playerUiSettings,
    setWaveformSpreadMode,
  } from '@/stores/playback/playerSettings.svelte';

  let open = $state(false);
  let rootEl = $state<HTMLDivElement | null>(null);

  function toggleMenu() {
    open = !open;
  }

  function closeMenu() {
    open = false;
  }

  function selectMode(id: (typeof WAVEFORM_SPREAD_OPTIONS)[number]['id']) {
    setWaveformSpreadMode(id);
    closeMenu();
  }

  $effect(() => {
    if (!open) return;

    const onPointerDown = (event: PointerEvent) => {
      const target = event.target;
      if (target instanceof Node && rootEl?.contains(target)) {
        return;
      }
      closeMenu();
    };

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        closeMenu();
      }
    };

    document.addEventListener('pointerdown', onPointerDown, true);
    window.addEventListener('keydown', onKeyDown);
    return () => {
      document.removeEventListener('pointerdown', onPointerDown, true);
      window.removeEventListener('keydown', onKeyDown);
    };
  });
</script>

<div class="settings-menu" bind:this={rootEl}>
  <button
    type="button"
    class="settings-trigger"
    class:open
    aria-label="界面设置"
    aria-expanded={open}
    aria-haspopup="dialog"
    onclick={toggleMenu}
  >
    <Settings size={16} strokeWidth={1.75} />
    <span>界面设置</span>
    <span class="chevron" class:open>
      <ChevronDown size={14} />
    </span>
  </button>

  {#if open}
    <div class="settings-popover" role="dialog" aria-label="界面设置">
      <p class="popover-heading">波形动画</p>
      <p class="popover-hint">封面下方实时波形的扩散方向</p>
      <ul class="option-list" role="radiogroup" aria-label="波形扩散方向">
        {#each WAVEFORM_SPREAD_OPTIONS as option (option.id)}
          <li>
            <button
              type="button"
              class="option-btn"
              class:selected={playerUiSettings.waveformSpread === option.id}
              role="radio"
              aria-checked={playerUiSettings.waveformSpread === option.id}
              onclick={() => selectMode(option.id)}
            >
              <span class="option-text">
                <span class="option-label">{option.label}</span>
                <span class="option-desc">{option.description}</span>
              </span>
              {#if playerUiSettings.waveformSpread === option.id}
                <Check size={16} strokeWidth={2.25} class="check-icon" />
              {/if}
            </button>
          </li>
        {/each}
      </ul>
    </div>
  {/if}
</div>

<style>
  .settings-menu {
    position: relative;
  }

  .settings-trigger {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    border: none;
    border-radius: 8px;
    background: transparent;
    color: rgba(255, 255, 255, 0.55);
    font-size: 0.8125rem;
    font-family: inherit;
    cursor: pointer;
    transition: background 0.15s ease, color 0.15s ease;
  }

  .settings-trigger:hover,
  .settings-trigger.open {
    color: rgba(255, 255, 255, 0.9);
    background: rgba(255, 255, 255, 0.08);
  }

  .chevron {
    display: inline-flex;
    transition: transform 0.15s ease;
  }

  .chevron.open {
    transform: rotate(180deg);
  }

  .settings-popover {
    position: absolute;
    top: calc(100% + 8px);
    right: 0;
    z-index: 20;
    width: min(280px, calc(100vw - 40px));
    padding: 14px;
    border-radius: 12px;
    border: 1px solid rgba(255, 255, 255, 0.12);
    background: rgba(18, 14, 12, 0.94);
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.45);
    backdrop-filter: blur(16px);
  }

  .popover-heading {
    margin: 0;
    font-size: 13px;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.92);
  }

  .popover-hint {
    margin: 4px 0 12px;
    font-size: 11px;
    color: rgba(255, 255, 255, 0.42);
  }

  .option-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .option-btn {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    padding: 10px 12px;
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.04);
    color: inherit;
    font-family: inherit;
    text-align: left;
    cursor: pointer;
    transition: border-color 0.15s ease, background 0.15s ease;
  }

  .option-btn:hover {
    background: rgba(255, 255, 255, 0.08);
    border-color: rgba(255, 255, 255, 0.14);
  }

  .option-btn.selected {
    border-color: rgba(255, 255, 255, 0.28);
    background: rgba(255, 255, 255, 0.1);
  }

  .option-text {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .option-label {
    font-size: 13px;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.9);
  }

  .option-desc {
    font-size: 11px;
    color: rgba(255, 255, 255, 0.45);
  }

  .option-btn :global(.check-icon) {
    flex-shrink: 0;
    color: rgba(255, 255, 255, 0.88);
  }
</style>
