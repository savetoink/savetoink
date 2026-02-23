<script lang="ts">
	import type { PageData } from './$types';
	import ArticleControls from '$lib/components/ArticleControls.svelte';
	import ArticleMetaAccordion from '$lib/components/ArticleMetaAccordion.svelte';

	let { data }: { data: PageData } = $props();
	const title = $derived(data.title || data.url);
</script>

<article>
	<header>
		<h1>
			<a
				href={data.url}
				target="_blank"
				rel="external noreferrer"
				title="Open the original link"
				data-tooltip="Open the original link">{title}</a
			>
		</h1>

		<ArticleMetaAccordion article={data} />
	</header>
	<section>
		<!-- eslint-disable-next-line svelte/no-at-html-tags -->
		{@html data.content}
	</section>
	<footer>
		<ArticleControls article={data} />
	</footer>
</article>
