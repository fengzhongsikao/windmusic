/** SPA 路由表：路径 → 页面组件 */
import DiscoverPage from '@/pages/discover/DiscoverPage.svelte';
import RecommendPage from '@/pages/browse/RecommendPage.svelte';
import RankingPage from '@/pages/browse/RankingPage.svelte';
import FavoritesPage from '@/pages/library/FavoritesPage.svelte';
import RecentPage from '@/pages/library/RecentPage.svelte';
import LocalPage from '@/pages/library/LocalPage.svelte';
import SettingsPage from '@/pages/settings/SettingsPage.svelte';
import SearchPage from '@/pages/search/SearchPage.svelte';

export default {
  '/': DiscoverPage,
  '/discover': DiscoverPage,
  '/recommend': RecommendPage,
  '/ranking': RankingPage,
  '/favorites': FavoritesPage,
  '/recent': RecentPage,
  '/local': LocalPage,
  '/search': SearchPage,
  '/settings': SettingsPage,
};
