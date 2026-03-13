import { defineConfig } from 'vitest/config';
import { sveltekit } from '@sveltejs/kit/vite';
import { getBuildConfig } from '@savetoink/shared/lib/build-config';
import path from 'path';

const { version } = getBuildConfig();

export default defineConfig({
	plugins: [sveltekit()],
	ssr: {
		external: ['@sentry/sveltekit']
	},
	define: {
		__APP_VERSION__: JSON.stringify(version)
	},
	resolve: {
		alias: {
			'$app/stores': path.resolve('./src/stubs/app-stores.ts')
		}
	}
});
