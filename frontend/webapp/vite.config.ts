import { sentrySvelteKit } from '@sentry/sveltekit';
import { defineConfig } from 'vitest/config';
import { playwright } from '@vitest/browser-playwright';
import { sveltekit } from '@sveltejs/kit/vite';
import { getBuildConfig } from '@savetoink/shared/lib/build-config';

const { version, buildDate, gitHash } = getBuildConfig();

export default defineConfig({
	plugins: [
		sentrySvelteKit({
			org: process.env.PUBLIC_SENTRY_TEAM,
			project: process.env.PUBLIC_SENTRY_PROJECT,
			adapter: 'cloudflare'
		}),
		sveltekit()
	],
	define: {
		__APP_VERSION__: JSON.stringify(version),
		__BUILD_DATE__: JSON.stringify(buildDate),
		__GIT_HASH__: JSON.stringify(gitHash)
	},
	test: {
		expect: { requireAssertions: true },
		projects: [
			{
				extends: './vite.config.ts',
				test: {
					name: 'client',
					browser: {
						enabled: true,
						provider: playwright(),
						instances: [{ browser: 'chromium', headless: true }]
					},
					include: ['src/**/*.svelte.{test,spec}.{js,ts}'],
					exclude: ['src/lib/server/**']
				}
			},

			{
				extends: './vite.config.ts',
				test: {
					name: 'server',
					environment: 'node',
					include: ['src/**/*.{test,spec}.{js,ts}'],
					exclude: ['src/**/*.svelte.{test,spec}.{js,ts}']
				}
			}
		]
	}
});
