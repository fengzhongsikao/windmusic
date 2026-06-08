export function parseSearchParams(querystring: string | undefined) {
  const params = new URLSearchParams(querystring ?? '');
  const page = parseInt(params.get('page') ?? '1', 10);
  return {
    q: params.get('q')?.trim() ?? '',
    page: Number.isFinite(page) && page > 0 ? page : 1,
  };
}

export function buildSearchHref(keyword: string, page = 1) {
  const params = new URLSearchParams();
  const q = keyword.trim();
  if (q) {
    params.set('q', q);
  }
  if (page > 1) {
    params.set('page', String(page));
  }
  const qs = params.toString();
  return qs ? `/search?${qs}` : '/search';
}
