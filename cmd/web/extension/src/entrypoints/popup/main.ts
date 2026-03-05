import { mount } from 'svelte';
import App from './App.svelte';
import '@savetoink/shared/css';

if (import.meta.env.DEV) {
	import('@savetoink/shared/css-dev');
}

const app = mount(App, {
	target: document.getElementById('app')!
});

export default app;
