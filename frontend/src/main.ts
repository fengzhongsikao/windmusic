import '@/style.css';
import App from '@/App.svelte';
import { mount, unmount } from 'svelte';

const target = document.getElementById('app')!;

const app = mount(App, {
  target,
});

if (import.meta.hot) {
  import.meta.hot.dispose(() => {
    unmount(app);
  });
}

export default app;
