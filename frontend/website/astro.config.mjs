import { copyFileSync, mkdirSync } from 'node:fs';
import { defineConfig } from 'astro/config';
import sitemap from '@astrojs/sitemap';

const syncOpenApiSpec = {
	name: 'sync-openapi-spec',
	hooks: {
		'astro:config:setup': ({ logger }) => {
			mkdirSync('public', { recursive: true });
			copyFileSync('../../backend/lib/server/handlers/openapi.yaml', 'public/openapi.yaml');
			logger.info('OpenAPI spec synced to public/');
		}
	}
};

export default defineConfig({
	site: 'https://www.saveto.ink',
	output: 'static',
	integrations: [syncOpenApiSpec, sitemap()]
});
