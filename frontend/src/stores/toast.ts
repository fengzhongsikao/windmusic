import { writable } from 'svelte/store';

export type ToastType = 'error' | 'success' | 'info';

export type ToastItem = {
  id: number;
  message: string;
  type: ToastType;
};

const toastsStore = writable<ToastItem[]>([]);
let idSeed = 0;

export const toasts = toastsStore;

function push(message: string, type: ToastType, timeoutMs = 2200) {
  const id = ++idSeed;
  toastsStore.update((items) => [...items, { id, message, type }]);
  if (timeoutMs > 0) {
    window.setTimeout(() => dismiss(id), timeoutMs);
  }
}

export function error(message: string) {
  push(message, 'error');
}

export function dismiss(id: number) {
  toastsStore.update((items) => items.filter((item) => item.id !== id));
}
