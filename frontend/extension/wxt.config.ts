import path from 'path';
import { defineConfig } from 'wxt';
import { getBuildConfig } from '@savetoink/shared/lib/build-config';

// See https://wxt.dev/api/config.html
export default defineConfig({
	srcDir: 'src',
	outDir: 'dist',
	manifest: {
		permissions: ['activeTab', 'contextMenus'],
		action: {
			default_title: 'Save to Ink'
		},
		browser_specific_settings: {
			gecko: {
				data_collection_permissions: {
					required: ['none']
				}
			} as Record<string, unknown>
		}
	},
	vite() {
		const { version } = getBuildConfig();
		return {
			envPrefix: 'PUBLIC_',
			define: {
				__APP_VERSION__: JSON.stringify(version)
			},
			resolve: {
				alias: {
					'@savetoink/shared': path.resolve(__dirname, '../shared/src')
				}
			}
		};
	}
});
