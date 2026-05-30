import { ListRecent } from '../../../wailsjs/go/main/App';
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import { music } from '../../../wailsjs/go/models';
import { normalizeRecentSong, type RecentSong } from '@/lib/recentSong';

export const RECENT_UPDATED_EVENT = 'recent:updated';

export const recentState = $state({
  items: [] as RecentSong[],
  loaded: false,
});

let syncInitialized = false;

function applyRecent(raw: music.RecentSong[]) {
  recentState.items = (raw ?? []).map((item) => normalizeRecentSong(item));
  recentState.loaded = true;
}

export function initRecentSync(): () => void {
  if (syncInitialized) {
    return () => {};
  }
  syncInitialized = true;

  const offUpdated = EventsOn(RECENT_UPDATED_EVENT, (payload: music.RecentSong[]) => {
    applyRecent(payload ?? []);
  });

  void ListRecent()
    .then((items) => applyRecent(items ?? []))
    .catch(() => {});

  return () => {
    offUpdated();
    syncInitialized = false;
  };
}

export function getRecentSongs(): RecentSong[] {
  return recentState.items;
}
