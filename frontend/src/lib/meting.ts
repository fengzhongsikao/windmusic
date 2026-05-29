const METING_URLS_KEY = 'windmusic:meting-urls';
const METING_URL_KEY_LEGACY = 'windmusic:meting-url';
const METING_ACTIVE_KEY = 'windmusic:meting-active';
const METING_PLATFORM_KEY = 'windmusic:meting-platform';

export type MetingPlatform = 'netease';

/** Meting API（如 meting.mikus.ink）仅 netease 支持 type=search。 */
const DEFAULT_METING_PLATFORM: MetingPlatform = 'netease';

function normalizeMetingPlatform(raw: string): MetingPlatform {
  const v = raw.trim().toLowerCase();
  if (v === 'netease' || v === 'wy' || v === '163') {
    return 'netease';
  }
  return 'netease';
}

export function normalizeMetingURL(url: string): string {
  return url.trim().replace(/\/+$/, '');
}

function readMetingURLs(): string[] {
  if (typeof window === 'undefined') return [];

  const raw = window.localStorage.getItem(METING_URLS_KEY);
  if (raw) {
    try {
      const parsed = JSON.parse(raw) as unknown;
      if (Array.isArray(parsed)) {
        const urls: string[] = [];
        const seen = new Set<string>();
        for (const item of parsed) {
          const normalized = normalizeMetingURL(String(item ?? ''));
          if (!normalized || seen.has(normalized)) continue;
          seen.add(normalized);
          urls.push(normalized);
        }
        return urls;
      }
    } catch {
      // fall through to legacy migration
    }
  }

  const legacy = normalizeMetingURL(window.localStorage.getItem(METING_URL_KEY_LEGACY) ?? '');
  if (!legacy) return [];

  writeMetingURLs([legacy]);
  window.localStorage.removeItem(METING_URL_KEY_LEGACY);
  return [legacy];
}

function writeMetingURLs(urls: string[]) {
  if (typeof window === 'undefined') return;
  if (urls.length === 0) {
    window.localStorage.removeItem(METING_URLS_KEY);
    window.localStorage.removeItem(METING_ACTIVE_KEY);
    return;
  }
  window.localStorage.setItem(METING_URLS_KEY, JSON.stringify(urls));
  const active = normalizeMetingURL(window.localStorage.getItem(METING_ACTIVE_KEY) ?? '');
  if (!active || !urls.includes(active)) {
    window.localStorage.setItem(METING_ACTIVE_KEY, urls[0]);
  }
}

export function getMetingURLs(): string[] {
  return readMetingURLs();
}

/** 当前用于搜索/播放的 Meting 节点地址。 */
export function getMetingURL(): string {
  const urls = readMetingURLs();
  if (urls.length === 0) return '';
  const active = normalizeMetingURL(
    typeof window === 'undefined' ? '' : (window.localStorage.getItem(METING_ACTIVE_KEY) ?? ''),
  );
  if (active && urls.includes(active)) return active;
  return urls[0];
}

export function getActiveMetingURL(): string {
  return getMetingURL();
}

export function setActiveMetingURL(url: string) {
  if (typeof window === 'undefined') return;
  const normalized = normalizeMetingURL(url);
  const urls = readMetingURLs();
  if (!normalized || !urls.includes(normalized)) return;
  window.localStorage.setItem(METING_ACTIVE_KEY, normalized);
}

export function addMetingURL(url: string): string | null {
  const normalized = normalizeMetingURL(url);
  if (!normalized) return '请输入 Meting 地址';
  if (!/^https?:\/\//i.test(normalized)) return '地址需以 http:// 或 https:// 开头';

  const urls = readMetingURLs();
  if (urls.includes(normalized)) return '该地址已存在';

  writeMetingURLs([...urls, normalized]);
  return null;
}

export function removeMetingURL(url: string) {
  const normalized = normalizeMetingURL(url);
  const urls = readMetingURLs().filter((item) => item !== normalized);
  writeMetingURLs(urls);
}

export function getMetingPlatform(): MetingPlatform {
  if (typeof window === 'undefined') return DEFAULT_METING_PLATFORM;
  const raw = window.localStorage.getItem(METING_PLATFORM_KEY) ?? '';
  return normalizeMetingPlatform(raw);
}

export function setMetingPlatform(platform: string) {
  if (typeof window === 'undefined') return;
  window.localStorage.setItem(METING_PLATFORM_KEY, normalizeMetingPlatform(platform));
}

export function metingSourceId(url: string): string {
  return `meting::${normalizeMetingURL(url)}`;
}
