import { initLocalLibrarySync } from '@/stores/localLibrary.svelte';
import { initPlayerSettingsSync } from '@/stores/playerSettings.svelte';
import { initMetingSync } from '@/stores/meting.svelte';
import { initFavoritesSync } from '@/stores/favorites.svelte';
import { initPlaylistsSync } from '@/stores/playlistsStore.svelte';
import { initRecentSync } from '@/stores/recent.svelte';
import { clearLegacyClientStorage } from '@/lib/clearClientStorage';

export function initAppDataSync(): () => void {
  const stops = [
    initLocalLibrarySync(),
    initPlayerSettingsSync(),
    initMetingSync(),
    initFavoritesSync(),
    initPlaylistsSync(),
    initRecentSync(),
  ];

  clearLegacyClientStorage();

  return () => {
    for (const stop of stops) {
      stop();
    }
  };
}
