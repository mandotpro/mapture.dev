import { mount } from 'svelte';
import App from './App.svelte';
import './app.css';
import '@xyflow/svelte/dist/style.css';

const app = mount(App, {
  target: document.getElementById('app')!,
});

export default app;
