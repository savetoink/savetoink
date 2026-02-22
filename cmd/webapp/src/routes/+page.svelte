<script lang="ts">
	import ArticleMeta from '$lib/components/ArticleMeta.svelte';
	import Navigator from '$lib/components/Navigator.svelte';
	import type { PageData } from './$types';
	let { data }: { data: PageData } = $props();
</script>

<section aria-labelledby="my-list-heading">
	<hgroup>
		<h1 id="my-list-heading">My List</h1>
		<p>{data.total} articles</p>
	</hgroup>

	<ul>
		{#each data.articles as article (article.id)}
			<li>
				<ArticleMeta {article} mode="card" />
			</li>
		{/each}
	</ul>

	{#if data.articles.length > 0}
		<Navigator page={data.page} has_more={data.has_more} />
	{/if}
</section>
