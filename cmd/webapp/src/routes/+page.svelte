<script lang="ts">
	import ArticleCard from '$lib/components/ArticleCard.svelte';
	import Navigator from '$lib/components/Navigator.svelte';
	import type { PageData } from './$types';
	let { data }: { data: PageData } = $props();
</script>

<h1>My List (<span>{data.total} articles</span>)</h1>

<ul>
	{#each data.articles as article (article.id)}
		<li>
			<ArticleCard {article} />
		</li>
	{/each}
</ul>

<p>total: {data.total}</p>

{#if data.articles.length > 0}
	<Navigator page={data.page} has_more={data.has_more} />
{/if}
