/** 类型与 API 重导出（Runes 实现在 player.svelte.ts） */
export type { ViewMode, PlaybackContext, PlayerTrack, RepeatMode } from './player.svelte';

export {
  player,
  openImmersiveView,
  closeImmersiveView,
  toggleImmersiveView,
  setCurrentTrack,
  setQueue,
  playQueueTrack,
  removeQueueTrack,
  clearQueue,
  togglePlayByTrack,
  playAllTracks,
  togglePlayerPlayback,
  setPlaying,
  playNextTrack,
  playPreviousTrack,
  toggleShuffleMode,
  cycleRepeatMode,
  isCurrentTrack,
  repeatModeLabel,
  setPlayerVolume,
  togglePlayerMuted,
} from './player.svelte';
