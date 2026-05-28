/** 类型与 API 重导出（Runes 实现在 player.svelte.ts） */
export type { ViewMode, PlaybackContext, PlayerTrack } from './player.svelte';

export {
  player,
  openImmersiveView,
  closeImmersiveView,
  toggleImmersiveView,
  setCurrentTrack,
  togglePlayByTrack,
  togglePlayerPlayback,
  setPlaying,
} from './player.svelte';
