import { writable } from 'svelte/store';

export const songDetailOpen = writable(false);

export function openSongDetailDrawer() {
  songDetailOpen.set(true);
}

export function closeSongDetailDrawer() {
  songDetailOpen.set(false);
}

export function toggleSongDetailDrawer() {
  songDetailOpen.update((open) => !open);
}
