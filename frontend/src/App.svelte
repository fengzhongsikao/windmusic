<script lang="ts">
  import Router from 'svelte-spa-router';
  import Sidebar from '@/pages/layout/Sidebar.svelte';
  import PlayerBar from '@/pages/layout/PlayerBar.svelte';
  import SearchBar from '@/components/SearchBar.svelte';
  import routes from '@/routes';
  import { initAudioEngine } from '@/stores/audioEngine';
  import SongDetailDrawer from '@/components/SongDetailDrawer.svelte';

  let audioRoot = $state<HTMLDivElement | null>(null);
  let audioEl = $state<HTMLAudioElement | null>(null);

  $effect(() => {
    if (audioEl && audioRoot) {
      initAudioEngine(audioEl, audioRoot);
    }
  });
</script>

<div class="app-layout">
  <div class="app-body">
    <div class="app-sidebar">
      <Sidebar />
    </div>
    <div class="app-main">
      <div class="main-search">
        <SearchBar />
      </div>
      <div class="main-routes">
        <Router {routes} />
      </div>
    </div>
  </div>

  <div class="app-footer">
    <PlayerBar />
  </div>

  <div id="audio-root" class="audio-root" bind:this={audioRoot} aria-hidden="true">
    <audio bind:this={audioEl} preload="metadata"></audio>
  </div>

  <SongDetailDrawer />
</div>

<style>
  .app-layout {
    display: flex;
    flex-direction: column;
    height: 100vh;
    width: 100vw;
    overflow: hidden;
    background: #fff;
  }

  .app-body {
    flex: 1;
    display: flex;
    overflow: hidden;
  }

  .app-sidebar {
    width: 220px;
    flex-shrink: 0;
    overflow: hidden;
  }

  .app-main {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: #fff;
  }

  .main-search {
    flex-shrink: 0;
    padding: 16px 32px 0;
    -webkit-app-region: no-drag;
  }

  .main-routes {
    flex: 1;
    overflow-y: auto;
    padding: 16px 32px 24px;
  }

  .app-footer {
    height: 80px;
    flex-shrink: 0;
  }

  .audio-root {
    position: absolute;
    width: 0;
    height: 0;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
  }
</style>
