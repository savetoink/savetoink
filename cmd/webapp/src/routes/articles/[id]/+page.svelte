<script lang="ts">
	import type { PageData } from './$types';
	import ArticleControls from '$lib/components/ArticleControls.svelte';
	import ArticleMetaAccordion from '$lib/components/ArticleMetaAccordion.svelte';

	let { data }: { data: PageData } = $props();
</script>

<article>
	<header>
		{#if data.title}
			<h1>{data.title}</h1>
		{:else}
			<h1>{data.url}</h1>
		{/if}
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
