import { writable } from 'svelte/store';
import type { TrackItem } from '@/lib/track';
import defaultCover from '@/assets/images/default.jpg';

export type PlayerTrack = Pick<TrackItem, 'id' | 'title' | 'artist' | 'album' | 'coverUrl'>;

type PlayerState = {
  currentTrack: PlayerTrack;
  isPlaying: boolean;
};

const initialState: PlayerState = {
  currentTrack: {
    id: 'init-galaxy',
    title: '在银河中孤独摇摆',
    artist: '知更鸟 / HOYO-MiX',
    album: '未知专辑',
    coverUrl: defaultCover,
  },
  isPlaying: false,
};

export const playerState = writable<PlayerState>(initialState);

export function togglePlayByTrack(track: PlayerTrack) {
  playerState.update((state) => {
    if (String(state.currentTrack.id) === String(track.id)) {
      return { ...state, isPlaying: !state.isPlaying };
    }
    return {
      currentTrack: track,
      isPlaying: true,
    };
  });
}

export function togglePlayerPlayback() {
  playerState.update((state) => ({ ...state, isPlaying: !state.isPlaying }));
}
