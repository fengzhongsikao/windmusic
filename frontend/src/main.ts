import '@/style.css'
import App from '@/App.svelte'
import { mount, unmount } from 'svelte';
import { startAudioSync } from '@/stores/audioEngine';

startAudioSync();

const app = mount(App, { 
  target: document.getElementById("app")!,
});

export default app
