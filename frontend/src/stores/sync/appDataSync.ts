import { initLocalLibrarySync } from '@/stores/library/localLibrary.svelte';
import { initPlayerSettingsSync } from '@/stores/playback/playerSettings.svelte';
import { initMetingSync } from '@/stores/sources/meting.svelte';
import { initFavoritesSync } from '@/stores/library/favorites.svelte';
import { initPlaylistsSync } from '@/stores/library/playlistsStore.svelte';
import { initRecentSync } from '@/stores/library/recent.svelte';
import { clearLegacyClientStorage } from '@/lib/clearClientStorage';

let stopSync: (() => void) | null = null;

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

/** 尽早拉取应用数据，避免等 App onMount 才订阅后端事件 */
export function ensureAppDataSync(): () => void {
  if (stopSync) {
    return stopSync;
  }
  stopSync = initAppDataSync();
  return stopSync;
}

ensureAppDataSync();
