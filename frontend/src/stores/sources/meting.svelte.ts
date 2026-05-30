import {
  AddMetingURL,
  GetMetingSettings,
  RemoveMetingURL,
  SetActiveMetingURL,
  SetMetingPlatform,
} from '../../../wailsjs/go/main/App';
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import { music } from '../../../wailsjs/go/models';

export type MetingPlatform = 'netease';

export const METING_SETTINGS_UPDATED_EVENT = 'meting-settings:updated';

export const metingSettings = $state({
  urls: [] as string[],
  activeUrl: '',
  platform: 'netease' as MetingPlatform,
  loaded: false,
});

let syncInitialized = false;

export function normalizeMetingURL(url: string): string {
  return url.trim().replace(/\/+$/, '');
}

function normalizeMetingPlatform(raw: string): MetingPlatform {
  const v = raw.trim().toLowerCase();
  if (v === 'netease' || v === 'wy' || v === '163') {
    return 'netease';
  }
  return 'netease';
}

function applyMetingSettings(raw: music.MetingSettings) {
  metingSettings.urls = raw.urls ?? [];
  metingSettings.activeUrl = normalizeMetingURL(raw.activeUrl ?? '');
  metingSettings.platform = normalizeMetingPlatform(raw.platform ?? 'netease');
  metingSettings.loaded = true;
}

export function initMetingSync(): () => void {
  if (syncInitialized) {
    return () => {};
  }
  syncInitialized = true;

  const offUpdated = EventsOn(METING_SETTINGS_UPDATED_EVENT, (payload: music.MetingSettings) => {
    applyMetingSettings(payload);
  });

  void GetMetingSettings()
    .then((settings) => applyMetingSettings(settings))
    .catch(() => {});

  return () => {
    offUpdated();
    syncInitialized = false;
  };
}

export function getMetingURLs(): string[] {
  return metingSettings.urls;
}

export function getMetingURL(): string {
  if (metingSettings.urls.length === 0) {
    return '';
  }
  const active = normalizeMetingURL(metingSettings.activeUrl);
  if (active && metingSettings.urls.includes(active)) {
    return active;
  }
  return metingSettings.urls[0];
}

export function getActiveMetingURL(): string {
  return getMetingURL();
}

export async function setActiveMetingURL(url: string): Promise<void> {
  const normalized = normalizeMetingURL(url);
  if (!normalized || !metingSettings.urls.includes(normalized)) {
    return;
  }
  await SetActiveMetingURL(normalized);
}

export async function addMetingURL(url: string): Promise<string | null> {
  const normalized = normalizeMetingURL(url);
  if (!normalized) {
    return '请输入 Meting 地址';
  }
  if (!/^https?:\/\//i.test(normalized)) {
    return '地址需以 http:// 或 https:// 开头';
  }
  if (metingSettings.urls.includes(normalized)) {
    return '该地址已存在';
  }
  try {
    await AddMetingURL(normalized);
    return null;
  } catch (err) {
    return err instanceof Error ? err.message : String(err);
  }
}

export async function removeMetingURL(url: string): Promise<void> {
  await RemoveMetingURL(normalizeMetingURL(url));
}

export function getMetingPlatform(): MetingPlatform {
  return metingSettings.platform;
}

export async function setMetingPlatform(platform: string): Promise<void> {
  await SetMetingPlatform(platform);
}

export function metingSourceId(url: string): string {
  return `meting::${normalizeMetingURL(url)}`;
}
