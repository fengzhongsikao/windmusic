import { ListFavorites } from '../../../wailsjs/go/main/App';
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import { music } from '../../../wailsjs/go/models';
import { favoriteSongKey, normalizeFavoriteSong, type FavoriteSong } from '@/lib/library/favoriteSong';

export const FAVORITES_UPDATED_EVENT = 'favorites:updated';

export const favoritesState = $state({
  items: [] as FavoriteSong[],
  loaded: false,
});

let syncInitialized = false;

function applyFavorites(raw: music.FavoriteSong[]) {
  const seen = new Set<string>();
  favoritesState.items = raw
    .map((item) => normalizeFavoriteSong(item))
    .filter((item) => {
      const key = favoriteSongKey(item);
      if (seen.has(key)) {
        return false;
      }
      seen.add(key);
      return true;
    });
  favoritesState.loaded = true;
}

export function initFavoritesSync(): () => void {
  if (syncInitialized) {
    return () => {};
  }
  syncInitialized = true;

  const offUpdated = EventsOn(FAVORITES_UPDATED_EVENT, (payload: music.FavoriteSong[]) => {
    applyFavorites(payload ?? []);
  });

  void ListFavorites()
    .then((items) => applyFavorites(items ?? []))
    .catch(() => {});

  return () => {
    offUpdated();
    syncInitialized = false;
  };
}

export function getFavorites(): FavoriteSong[] {
  return favoritesState.items;
}

/** 从 Go 重新拉取收藏列表（添加/删除后或进入收藏页时调用） */
export async function refreshFavoritesFromBackend(): Promise<void> {
  try {
    const items = await ListFavorites();
    applyFavorites(items ?? []);
  } catch {
    // 保留现有内存状态
  }
}
