import { mount } from 'svelte';
import App from './App.svelte';
import { getCSS } from '@savetoink/shared/css';

getCSS(import.meta.env.DEV);

const app = mount(App, {
	target: document.getElementById('app')!
});

export default app;
