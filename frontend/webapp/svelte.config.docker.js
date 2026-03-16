import adapter from 'svelte-adapter-bun';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		adapter: adapter({
			precompress: false
		}),
		alias: {
			'@savetoink/shared': '../shared/src'
		}
	}
};

export default config;
