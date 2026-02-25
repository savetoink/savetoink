<script lang="ts">
	import ArticleMetaItem from '$lib/components/ArticleMetaItem.svelte';
	import Navigator from '$lib/components/Navigator.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<section>
	{#if data.articles.length === 0}
		<p>No articles found</p>
	{:else}
		<ul>
			{#each data.articles as article (article.id)}
				<ArticleMetaItem {article} />
			{/each}
		</ul>
	{/if}

	<Navigator page={data.page} has_more={data.has_more} />
</section>

<style>
	ul {
		padding-left: 0;
	}
</style>
