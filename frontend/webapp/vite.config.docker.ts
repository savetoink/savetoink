import { defineConfig } from 'vitest/config';
import { sveltekit } from '@sveltejs/kit/vite';
import { getBuildConfig } from '@savetoink/shared/lib/build-config';

const { version } = getBuildConfig();

export default defineConfig({
	plugins: [sveltekit()],
	define: {
		__APP_VERSION__: JSON.stringify(version)
	}
});
