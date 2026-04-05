import { sentrySvelteKit } from '@sentry/sveltekit';
import { defineConfig } from 'vitest/config';
import { sveltekit } from '@sveltejs/kit/vite';
import { getBuildConfig } from '@savetoink/shared/lib/build-config';

const { version } = getBuildConfig();

export default defineConfig({
	plugins: [
		sentrySvelteKit({
			org: process.env.PUBLIC_SENTRY_TEAM,
			project: process.env.PUBLIC_SENTRY_PROJECT,
			adapter: 'cloudflare',
			autoUploadSourceMaps: true,
			sourcemaps: {
				assets: [
					'./.svelte-kit/output/client/**/*.{js,js.map}',
					'./.svelte-kit/output/server/**/*.{js,js.map}'
				]
			},
			release: {
				name: version,
				inject: true
			}
		}),
		sveltekit()
	],
	define: {
		__APP_VERSION__: JSON.stringify(version)
	},
	test: {
		expect: { requireAssertions: true },
		include: ['src/**/*.{test,spec}.{js,ts}']
	}
});
