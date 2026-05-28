<script lang="ts">
  import { Slider } from '@skeletonlabs/skeleton-svelte';

  interface Props {
    value: number;
    max: number;
    step?: number;
    disabled?: boolean;
    ariaLabel?: string;
    oninput?: (value: number) => void;
    onchange?: (value: number) => void;
  }

  let {
    value,
    max,
    step = 0.1,
    disabled = false,
    ariaLabel = '播放进度',
    oninput,
    onchange,
  }: Props = $props();

  const safeMax = $derived(Math.max(max, step));
  const clampedValue = $derived(Math.min(safeMax, Math.max(0, value)));

  function handleValueChange(details: { value: number[] }) {
    const next = Number(details.value?.[0] ?? clampedValue);
    oninput?.(next);
  }

  function handleValueChangeEnd(details: { value: number[] }) {
    const next = Number(details.value?.[0] ?? clampedValue);
    onchange?.(next);
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="skeleton-slider progress-slider" onpointerdown={(e) => e.stopPropagation()}>
  <Slider
    value={[clampedValue]}
    min={0}
    max={safeMax}
    {step}
    {disabled}
    aria-label={[ariaLabel]}
    onValueChange={handleValueChange}
    onValueChangeEnd={handleValueChangeEnd}
  >
    <Slider.Control>
      <Slider.Track class="bg-primary-50-950">
        <Slider.Range class="bg-primary-500" />
      </Slider.Track>
      <Slider.Thumb index={0} class="ring-primary-500">
        <Slider.HiddenInput />
      </Slider.Thumb>
    </Slider.Control>
  </Slider>
</div>
