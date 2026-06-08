<script lang="ts">
  import { Settings, ChevronDown, Check } from '@lucide/svelte';
  import { Popover, Portal, Switch } from '@skeletonlabs/skeleton-svelte';
  import { WAVEFORM_SPREAD_OPTIONS } from '@/lib/playback/waveformSpread';
  import { DETAIL_COVER_SHAPE_OPTIONS } from '@/lib/playback/detailCoverShape';
  import {
    playerUiSettings,
    setDetailCoverShape,
    setDetailCoverSpin,
    setDetailHideLyrics,
    setDetailHidePlayerBar,
    setDetailHideVisual,
    setWaveformSpreadMode,
  } from '@/stores/playback/playerSettings.svelte';

  const LAYOUT_TOGGLES = [
    {
      id: 'hideLyrics' as const,
      label: '隐藏歌词区域',
      description: '左侧专辑与波形全屏显示',
    },
    {
      id: 'hideVisual' as const,
      label: '隐藏专辑区域',
      description: '右侧歌词全屏显示',
    },
    {
      id: 'hidePlayerBar' as const,
      label: '隐藏播放器控制栏',
      description: '详情页全屏显示，不保留底部控制条',
    },
  ];

  const COVER_SPIN_TOGGLE = {
    id: 'coverSpin' as const,
    label: '封面旋转',
    description: '播放时旋转圆形封面',
  };

  let open = $state(false);

  function selectMode(id: (typeof WAVEFORM_SPREAD_OPTIONS)[number]['id']) {
    setWaveformSpreadMode(id);
  }

  function selectCoverShape(id: (typeof DETAIL_COVER_SHAPE_OPTIONS)[number]['id']) {
    setDetailCoverShape(id);
  }

  function layoutChecked(id: (typeof LAYOUT_TOGGLES)[number]['id']): boolean {
    if (id === 'hideLyrics') return playerUiSettings.detailHideLyrics;
    if (id === 'hideVisual') return playerUiSettings.detailHideVisual;
    return playerUiSettings.detailHidePlayerBar;
  }

  function handleLayoutChange(id: (typeof LAYOUT_TOGGLES)[number]['id'], checked: boolean) {
    if (id === 'hideLyrics') {
      setDetailHideLyrics(checked);
      return;
    }
    if (id === 'hideVisual') {
      setDetailHideVisual(checked);
      return;
    }
    setDetailHidePlayerBar(checked);
  }
</script>

<div class="detail-settings-root">
<Popover
  {open}
  onOpenChange={(details) => {
    open = details.open;
  }}
  positioning={{ placement: 'bottom-end', gutter: 8 }}
>
  <Popover.Trigger
    class="btn preset-tonal detail-settings-trigger"
    aria-label="界面设置"
  >
    <Settings size={16} strokeWidth={1.75} />
    <span>界面设置</span>
    <span class="chevron" class:open>
      <ChevronDown size={14} />
    </span>
  </Popover.Trigger>

  <Portal>
    <Popover.Positioner>
      <Popover.Content class="card detail-settings-content">
        <p class="section-title">波形动画</p>
        <p class="section-hint">播放点始终在正中；切换模式可预览不同动画风格</p>

        <div class="option-stack" role="radiogroup" aria-label="波形扩散方向">
          {#each WAVEFORM_SPREAD_OPTIONS as option (option.id)}
            {@const selected = playerUiSettings.waveformSpread === option.id}
            <button
              type="button"
              class="btn w-full justify-between gap-3 text-left option-btn"
              class:preset-filled={selected}
              class:preset-tonal={!selected}
              role="radio"
              aria-checked={selected}
              onclick={() => selectMode(option.id)}
            >
              <span class="option-text">
                <span class="option-label">{option.label}</span>
                <span class="option-desc">{option.description}</span>
              </span>
              {#if selected}
                <Check size={16} strokeWidth={2.25} class="check-icon shrink-0" />
              {/if}
            </button>
          {/each}
        </div>

        <hr class="hr section-divider" />

        <p class="section-title">布局与封面</p>

        <p class="subsection-title">封面样式</p>
        <div class="option-stack" role="radiogroup" aria-label="封面样式">
          {#each DETAIL_COVER_SHAPE_OPTIONS as option (option.id)}
            {@const selected = playerUiSettings.detailCoverShape === option.id}
            <button
              type="button"
              class="btn w-full justify-between gap-3 text-left option-btn"
              class:preset-filled={selected}
              class:preset-tonal={!selected}
              role="radio"
              aria-checked={selected}
              onclick={() => selectCoverShape(option.id)}
            >
              <span class="option-text">
                <span class="option-label">{option.label}</span>
                <span class="option-desc">{option.description}</span>
              </span>
              {#if selected}
                <Check size={16} strokeWidth={2.25} class="check-icon shrink-0" />
              {/if}
            </button>
          {/each}
        </div>

        {#if playerUiSettings.detailCoverShape === 'round'}
          <div class="switch-stack cover-spin-stack" aria-label="圆形封面旋转">
            <Switch
              class="layout-switch"
              checked={playerUiSettings.detailCoverSpin}
              onCheckedChange={(details) => setDetailCoverSpin(details.checked)}
            >
              <span class="switch-copy">
                <Switch.Label class="switch-label">{COVER_SPIN_TOGGLE.label}</Switch.Label>
                <span class="switch-desc">{COVER_SPIN_TOGGLE.description}</span>
              </span>
              <Switch.Control class="switch-control">
                <Switch.Thumb />
              </Switch.Control>
              <Switch.HiddenInput />
            </Switch>
          </div>
        {/if}

        <hr class="hr section-divider" />

        <div class="switch-stack" aria-label="详情页布局">
          {#each LAYOUT_TOGGLES as toggle, index (toggle.id)}
            <Switch
              class="layout-switch"
              checked={layoutChecked(toggle.id)}
              onCheckedChange={(details) => handleLayoutChange(toggle.id, details.checked)}
            >
              <span class="switch-copy">
                <Switch.Label class="switch-label">{toggle.label}</Switch.Label>
                <span class="switch-desc">{toggle.description}</span>
              </span>
              <Switch.Control class="switch-control">
                <Switch.Thumb />
              </Switch.Control>
              <Switch.HiddenInput />
            </Switch>
            {#if index < LAYOUT_TOGGLES.length - 1}
              <hr class="hr" />
            {/if}
          {/each}
        </div>
      </Popover.Content>
    </Popover.Positioner>
  </Portal>
</Popover>
</div>

<style>
  .detail-settings-root {
    position: relative;
  }

  :global(.detail-settings-trigger) {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding-inline: 12px;
    min-height: 34px;
    font-size: 0.8125rem;
    color: rgba(255, 255, 255, 0.55);
    border: none;
    background: transparent;
  }

  :global(.detail-settings-trigger:hover),
  :global(.detail-settings-trigger[data-state='open']) {
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

  :global(.detail-settings-content) {
    width: min(300px, calc(100vw - 40px));
    padding: 14px;
    z-index: 200;
    background: rgba(18, 14, 12, 0.96);
    border-color: rgba(255, 255, 255, 0.12);
    color: rgba(255, 255, 255, 0.92);
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.45);
    backdrop-filter: blur(16px);
  }

  .section-title {
    margin: 0;
    font-size: 13px;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.92);
  }

  .section-hint {
    margin: 4px 0 12px;
    font-size: 11px;
    color: rgba(255, 255, 255, 0.42);
  }

  .section-divider {
    margin: 14px 0;
    border-color: rgba(255, 255, 255, 0.1);
  }

  .subsection-title {
    margin: 0 0 8px;
    font-size: 12px;
    font-weight: 500;
    color: rgba(255, 255, 255, 0.55);
  }

  .cover-spin-stack {
    margin-top: 10px;
  }

  .option-stack {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  :global(.option-btn) {
    min-height: auto;
    padding: 10px 12px;
    border: 1px solid rgba(255, 255, 255, 0.08);
  }

  :global(.option-btn.preset-tonal) {
    background: rgba(255, 255, 255, 0.04);
    color: inherit;
  }

  :global(.option-btn.preset-tonal:hover) {
    background: rgba(255, 255, 255, 0.08);
    border-color: rgba(255, 255, 255, 0.14);
  }

  :global(.option-btn.preset-filled) {
    border-color: rgba(255, 255, 255, 0.28);
    background: rgba(255, 255, 255, 0.12);
    color: rgba(255, 255, 255, 0.95);
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
  }

  .option-desc {
    font-size: 11px;
    opacity: 0.55;
  }

  :global(.check-icon) {
    color: rgba(255, 255, 255, 0.88);
  }

  .switch-stack :global(.hr) {
    border-color: rgba(255, 255, 255, 0.08);
  }

  :global(.layout-switch) {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    width: 100%;
    padding: 8px 2px;
  }

  .switch-copy {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
    flex: 1;
  }

  :global(.switch-label) {
    font-size: 13px;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.9);
  }

  .switch-desc {
    font-size: 11px;
    color: rgba(255, 255, 255, 0.45);
  }

  :global(.switch-control) {
    flex-shrink: 0;
    background: rgba(255, 255, 255, 0.16);
  }

  :global(.switch-control[data-state='checked']) {
    background: rgba(255, 255, 255, 0.42);
  }
</style>
