import { GetPlayerSettings, UpdatePlayerSettings } from '../../../wailsjs/go/main/App';
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import { music } from '../../../wailsjs/go/models';
import {
  DEFAULT_WAVEFORM_SPREAD,
  normalizeWaveformSpreadMode,
  type WaveformSpreadMode,
} from '@/lib/waveformSpread';
import {
  DEFAULT_DETAIL_COVER_SHAPE,
  normalizeDetailCoverShape,
  type DetailCoverShape,
} from '@/lib/detailCoverShape';
import { player, type RepeatMode } from '@/stores/playback/player.svelte';

export const PLAYER_SETTINGS_UPDATED_EVENT = 'player-settings:updated';

const DEFAULT_VOLUME = 30;

let syncInitialized = false;
let persistTimer: ReturnType<typeof setTimeout> | null = null;
let settingsTouchedLocally = false;

export const playerUiSettings = $state({
  waveformSpread: DEFAULT_WAVEFORM_SPREAD as WaveformSpreadMode,
  /** 切换波形模式时递增，供波形组件立即重绘（不持久化）。 */
  waveformSpreadRevision: 0,
  detailHideLyrics: false,
  detailHideVisual: false,
  detailCoverShape: DEFAULT_DETAIL_COVER_SHAPE as DetailCoverShape,
  detailCoverSpin: true,
  detailHidePlayerBar: false,
});

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
  playerUiSettings.waveformSpread = normalizeWaveformSpreadMode(raw.waveformSpread);
  playerUiSettings.detailHideLyrics = Boolean(raw.detailHideLyrics);
  playerUiSettings.detailHideVisual = Boolean(raw.detailHideVisual);
  if (playerUiSettings.detailHideLyrics && playerUiSettings.detailHideVisual) {
    playerUiSettings.detailHideVisual = false;
  }
  playerUiSettings.detailCoverShape = normalizeDetailCoverShape(raw.detailCoverShape);
  playerUiSettings.detailCoverSpin = raw.detailCoverSpin !== false;
  playerUiSettings.detailHidePlayerBar = Boolean(raw.detailHidePlayerBar);
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
    .then((settings) => {
      if (settingsTouchedLocally) {
        return;
      }
      applyPlayerSettings(settings);
    })
    .catch(() => {});

  return () => {
    offUpdated();
    syncInitialized = false;
  };
}

export function setWaveformSpreadMode(mode: WaveformSpreadMode) {
  if (playerUiSettings.waveformSpread === mode) {
    playerUiSettings.waveformSpreadRevision += 1;
    return;
  }
  settingsTouchedLocally = true;
  playerUiSettings.waveformSpread = mode;
  playerUiSettings.waveformSpreadRevision += 1;
  persistPlayerSettings();
}

export function setDetailHideLyrics(hide: boolean) {
  settingsTouchedLocally = true;
  playerUiSettings.detailHideLyrics = hide;
  if (hide && playerUiSettings.detailHideVisual) {
    playerUiSettings.detailHideVisual = false;
  }
  persistPlayerSettings();
}

export function setDetailHideVisual(hide: boolean) {
  settingsTouchedLocally = true;
  playerUiSettings.detailHideVisual = hide;
  if (hide && playerUiSettings.detailHideLyrics) {
    playerUiSettings.detailHideLyrics = false;
  }
  persistPlayerSettings();
}

export function setDetailCoverShape(shape: DetailCoverShape) {
  settingsTouchedLocally = true;
  playerUiSettings.detailCoverShape = shape;
  persistPlayerSettings();
}

export function setDetailCoverSpin(enabled: boolean) {
  settingsTouchedLocally = true;
  playerUiSettings.detailCoverSpin = enabled;
  persistPlayerSettings();
}

export function setDetailHidePlayerBar(hide: boolean) {
  settingsTouchedLocally = true;
  playerUiSettings.detailHidePlayerBar = hide;
  persistPlayerSettings();
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
      waveformSpread: playerUiSettings.waveformSpread,
      detailHideLyrics: playerUiSettings.detailHideLyrics,
      detailHideVisual: playerUiSettings.detailHideVisual,
      detailCoverShape: playerUiSettings.detailCoverShape,
      detailCoverSpin: playerUiSettings.detailCoverSpin,
      detailHidePlayerBar: playerUiSettings.detailHidePlayerBar,
    });
    void UpdatePlayerSettings(payload).catch(() => {});
  }, 120);
}
