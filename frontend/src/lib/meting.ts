const METING_URL_KEY = 'windmusic:meting-url';
const METING_PLATFORM_KEY = 'windmusic:meting-platform';

export function getMetingURL(): string {
  if (typeof window === 'undefined') return '';
  return (window.localStorage.getItem(METING_URL_KEY) ?? '').trim();
}

export function setMetingURL(url: string) {
  if (typeof window === 'undefined') return;
  const normalized = url.trim().replace(/\/+$/, '');
  if (!normalized) {
    window.localStorage.removeItem(METING_URL_KEY);
    return;
  }
  window.localStorage.setItem(METING_URL_KEY, normalized);
}

export function getMetingPlatform(): string {
  return 'tx';
}

export function setMetingPlatform(platform: string) {
  if (typeof window === 'undefined') return;
  window.localStorage.setItem(METING_PLATFORM_KEY, 'tx');
}

export function metingSourceId(url: string): string {
  return `meting::${url.trim().replace(/\/+$/, '')}`;
}

