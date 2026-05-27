import DiscoverPage from '@/pages/DiscoverPage.svelte';
import RecommendPage from '@/pages/RecommendPage.svelte';
import RankingPage from '@/pages/RankingPage.svelte';
import FavoritesPage from '@/pages/FavoritesPage.svelte';
import RecentPage from '@/pages/RecentPage.svelte';
import LocalPage from '@/pages/LocalPage.svelte';
import SettingsPage from '@/pages/SettingsPage.svelte';

export default {
  '/': DiscoverPage,
  '/discover': DiscoverPage,
  '/recommend': RecommendPage,
  '/ranking': RankingPage,
  '/favorites': FavoritesPage,
  '/recent': RecentPage,
  '/local': LocalPage,
  '/settings': SettingsPage,
};
