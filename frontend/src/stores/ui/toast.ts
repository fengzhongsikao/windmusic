import { createToaster } from '@skeletonlabs/skeleton-svelte';

export const toaster = createToaster({
  placement: 'top',
});

export function error(message: string) {
  toaster.error({ title: message });
}

export function success(message: string) {
  toaster.success({ title: message });
}

export function info(message: string) {
  toaster.info({ title: message });
}
