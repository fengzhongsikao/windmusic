<script lang="ts">
  import { player } from '@/stores/playback/player.svelte';
  import { syncPlayerState } from '@/stores/playback/audioEngine';

  $effect(() => {
    syncPlayerState(player.currentSong, player.isPlaying);
  });
</script>
