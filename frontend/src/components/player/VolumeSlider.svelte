<script lang="ts">
  import { Slider } from '@skeletonlabs/skeleton-svelte';

  interface Props {
    value: number;
    muted?: boolean;
    playing?: boolean;
    oninput?: (value: number) => void;
    onchange?: (value: number) => void;
  }

  let { value, muted = false, playing = false, oninput, onchange }: Props = $props();

  const displayValue = $derived(muted ? 0 : value);

  function handleValueChange(details: { value: number[] }) {
    const next = Number(details.value?.[0] ?? displayValue);
    oninput?.(next);
  }

  function handleValueChangeEnd(details: { value: number[] }) {
    const next = Number(details.value?.[0] ?? displayValue);
    onchange?.(next);
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="skeleton-slider volume-slider" onpointerdown={(e) => e.stopPropagation()}>
  <Slider
    value={[displayValue]}
    min={0}
    max={100}
    step={1}
    aria-label={['音量']}
    onValueChange={handleValueChange}
    onValueChangeEnd={handleValueChangeEnd}
  >
    <Slider.Control>
      <Slider.Track class="bg-secondary-50-950">
        <Slider.Range class="bg-secondary-500" />
      </Slider.Track>
      <Slider.Thumb index={0} class="ring-secondary-500">
        <Slider.HiddenInput />
      </Slider.Thumb>
    </Slider.Control>
  </Slider>
</div>
