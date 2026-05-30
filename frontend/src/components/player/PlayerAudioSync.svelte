<script lang="ts">
  import { player } from '@/stores/player.svelte';
  import { syncPlayerState } from '@/stores/audioEngine';

  $effect(() => {
    syncPlayerState(player.currentSong, player.isPlaying);
  });
</script>
