import { defineConfig } from 'astro/config';

// NOTICE: seems that >3.6.0 is buggy and won't build
// see https://github.com/savetoink/savetoink/issues/79
import sitemap from '@astrojs/sitemap';

export default defineConfig({
	site: 'https://www.saveto.ink',
	output: 'static',
	integrations: [sitemap()]
});
