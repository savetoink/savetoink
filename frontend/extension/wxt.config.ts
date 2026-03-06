import path from 'path';
import { defineConfig } from 'wxt';
import { getBuildConfig } from '@savetoink/shared/lib/build-config';

// See https://wxt.dev/api/config.html
export default defineConfig({
	srcDir: 'src',
	outDir: 'dist',
	modules: ['@wxt-dev/module-svelte'],
	manifest: {
		permissions: ['storage', 'tabs', 'contextMenus', 'identity']
	},
	vite() {
		const { version, buildDate, gitHash } = getBuildConfig();
		return {
			envPrefix: 'PUBLIC_',
			define: {
				__APP_VERSION__: JSON.stringify(version),
				__BUILD_DATE__: JSON.stringify(buildDate),
				__GIT_HASH__: JSON.stringify(gitHash)
			},
			resolve: {
				alias: {
					'@savetoink/shared': path.resolve(__dirname, '../shared/src')
				}
			}
		};
	}
});
