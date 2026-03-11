import { defineConfig } from 'astro/config';

import sitemap from '@astrojs/sitemap';

export default defineConfig({
	site: 'https://www.saveto.ink',
	output: 'static',
	integrations: [sitemap()]
});
