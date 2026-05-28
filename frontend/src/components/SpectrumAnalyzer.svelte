<script lang="ts">
  interface Props {
    active?: boolean;
    bars?: number;
    tone?: 'warm' | 'light';
  }

  let { active = false, bars = 48, tone = 'warm' }: Props = $props();

  const indices = $derived(Array.from({ length: bars }, (_, i) => i));
</script>

<div class="spectrum" class:live={active} class:light={tone === 'light'} aria-hidden="true">
  <div class="spectrum-bars">
    {#each indices as i (i)}
      <span class="bar" style="--i: {i}; --bars: {bars}"></span>
    {/each}
  </div>
  <div class="spectrum-reflect">
    {#each indices as i (i)}
      <span class="bar reflect" style="--i: {i}; --bars: {bars}"></span>
    {/each}
  </div>
</div>

<style>
  .spectrum {
    position: relative;
    width: 100%;
    max-width: 320px;
    height: 52px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: flex-end;
    overflow: hidden;
    mask-image: linear-gradient(90deg, transparent, #000 6%, #000 94%, transparent);
  }

  .spectrum-bars,
  .spectrum-reflect {
    display: flex;
    align-items: flex-end;
    justify-content: center;
    gap: 2px;
    width: 100%;
    height: 50%;
    padding: 0 2px;
  }

  .spectrum-reflect {
    align-items: flex-start;
    opacity: 0.2;
    transform: scaleY(-1);
  }

  .bar {
    flex: 1;
    max-width: 3px;
    min-width: 2px;
    height: 5px;
    border-radius: 1px;
    background: linear-gradient(
      180deg,
      rgba(255, 235, 210, 0.95) 0%,
      rgba(200, 150, 110, 0.75) 100%
    );
    transform-origin: bottom;
    opacity: 0.4;
    animation: spectrum-idle 1.3s ease-in-out infinite;
    animation-delay: calc((var(--i) / var(--bars)) * 0.75s);
    animation-play-state: paused;
  }

  .spectrum.light .bar {
    background: linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(255, 255, 255, 0.55) 100%);
    opacity: 0.5;
  }

  .spectrum.live .bar {
    animation-name: spectrum-live;
    animation-play-state: running;
    opacity: 0.95;
  }

  .spectrum.light.live .bar {
    opacity: 1;
  }

  @keyframes spectrum-idle {
    0%,
    100% {
      transform: scaleY(0.2);
    }
    50% {
      transform: scaleY(0.5);
    }
  }

  @keyframes spectrum-live {
    0%,
    100% {
      transform: scaleY(0.1);
    }
    50% {
      transform: scaleY(1);
    }
  }
</style>
