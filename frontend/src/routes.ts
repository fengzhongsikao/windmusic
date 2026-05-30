/** SPA 路由表：路径 → 页面组件（/local 由 App 层 keep-alive 单独挂载） */
import DiscoverPage from '@/pages/discover/DiscoverPage.svelte';
import RecommendPage from '@/pages/browse/RecommendPage.svelte';
import RankingPage from '@/pages/browse/RankingPage.svelte';
import FavoritesPage from '@/pages/library/FavoritesPage.svelte';
import RecentPage from '@/pages/library/RecentPage.svelte';
import PlaylistPage from '@/pages/library/PlaylistPage.svelte';
import SettingsPage from '@/pages/settings/SettingsPage.svelte';
import SearchPage from '@/pages/search/SearchPage.svelte';

export default {
  '/': DiscoverPage,
  '/discover': DiscoverPage,
  '/recommend': RecommendPage,
  '/ranking': RankingPage,
  '/favorites': FavoritesPage,
  '/recent': RecentPage,
  '/playlist/:id': PlaylistPage,
  '/search': SearchPage,
  '/settings': SettingsPage,
};
