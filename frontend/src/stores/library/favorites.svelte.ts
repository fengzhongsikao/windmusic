import { ListFavorites } from '../../../wailsjs/go/main/App';
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import { music } from '../../../wailsjs/go/models';
import { favoriteSongKey, normalizeFavoriteSong, type FavoriteSong } from '@/lib/favoriteSong';

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
