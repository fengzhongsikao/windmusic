import { GetPlayerSettings, UpdatePlayerSettings } from '../../../wailsjs/go/main/App';
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import { music } from '../../../wailsjs/go/models';
import { player, type RepeatMode } from '@/stores/playback/player.svelte';

export const PLAYER_SETTINGS_UPDATED_EVENT = 'player-settings:updated';

const DEFAULT_VOLUME = 30;

let syncInitialized = false;
let persistTimer: ReturnType<typeof setTimeout> | null = null;

function normalizeRepeatMode(value: string | undefined): RepeatMode {
  if (value === 'all' || value === 'one') {
    return value;
  }
  return 'off';
}

function applyPlayerSettings(raw: music.PlayerSettings) {
  const volume = Number.isFinite(raw.volume)
    ? Math.min(100, Math.max(0, Number(raw.volume)))
    : DEFAULT_VOLUME;
  player.volume = volume;
  player.isMuted = Boolean(raw.muted);
  player.repeatMode = normalizeRepeatMode(raw.repeatMode);
  player.isShuffled = Boolean(raw.shuffled);
}

export function initPlayerSettingsSync(): () => void {
  if (syncInitialized) {
    return () => {};
  }
  syncInitialized = true;

  const offUpdated = EventsOn(PLAYER_SETTINGS_UPDATED_EVENT, (payload: music.PlayerSettings) => {
    applyPlayerSettings(payload);
  });

  void GetPlayerSettings()
    .then((settings) => applyPlayerSettings(settings))
    .catch(() => {});

  return () => {
    offUpdated();
    syncInitialized = false;
  };
}

export function persistPlayerSettings() {
  if (persistTimer) {
    clearTimeout(persistTimer);
  }
  persistTimer = setTimeout(() => {
    persistTimer = null;
    const payload = music.PlayerSettings.createFrom({
      volume: player.volume,
      muted: player.isMuted,
      repeatMode: player.repeatMode,
      shuffled: player.isShuffled,
    });
    void UpdatePlayerSettings(payload).catch(() => {});
  }, 120);
}
