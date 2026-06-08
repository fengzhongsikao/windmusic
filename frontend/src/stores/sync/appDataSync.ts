import { waitForWailsBridge } from '@/lib/wails/wailsBridge';
import { initLocalLibrarySync } from '@/stores/library/localLibrary.svelte';
import { initPlayerSettingsSync } from '@/stores/playback/playerSettings.svelte';
import { initMetingSync } from '@/stores/sources/meting.svelte';
import { initFavoritesSync } from '@/stores/library/favorites.svelte';
import { initPlaylistsSync } from '@/stores/library/playlistsStore.svelte';
import { initRecentSync } from '@/stores/library/recent.svelte';
import { clearLegacyClientStorage } from '@/lib/clearClientStorage';

let stopSync: (() => void) | null = null;
let startPromise: Promise<void> | null = null;

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

/** 等 Wails 桥接就绪后再拉取应用数据，避免「Callback not registered」。 */
export function ensureAppDataSync(): () => void {
  if (stopSync) {
    return stopSync;
  }

  if (!startPromise) {
    startPromise = waitForWailsBridge().then((ready) => {
      if (!ready) {
        console.warn('[appDataSync] Wails 桥接未就绪，跳过数据同步初始化');
        return;
      }
      stopSync = initAppDataSync();
    });
  }

  return () => {
    stopSync?.();
    stopSync = null;
    startPromise = null;
  };
}

ensureAppDataSync();
