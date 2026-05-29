const LEGACY_KEYS = [
  'windmusic:player-settings',
  'windmusic:meting-urls',
  'windmusic:meting-url',
  'windmusic:meting-active',
  'windmusic:meting-platform',
] as const;

/** 清除曾用于 Web 前端的 localStorage 键（不再读取）。 */
export function clearLegacyClientStorage(): void {
  if (typeof window === 'undefined') {
    return;
  }
  for (const key of LEGACY_KEYS) {
    window.localStorage.removeItem(key);
  }
}
