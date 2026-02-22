<script lang="ts">
	import ArticleControls from '$lib/components/ArticleControls.svelte';
	import { resolve } from '$app/paths';
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
				<article>
					<header>
						{#if article.title}
							<h2><a href={resolve(`/articles/${article.id}`)}>{article.title}</a></h2>
						{:else}
							<h2><a href={resolve(`/articles/${article.id}`)}>{article.url}</a></h2>
						{/if}
						{#if article.imageUrl}
							<picture>
								<source srcset={article.imageUrl} type="image/webp" />
								<img src={article.imageUrl} alt={article.title} />
							</picture>
						{/if}
						{#if article.excerpt}
							<p>{article.excerpt}</p>
						{/if}
					</header>
					<footer>
						<ArticleControls {article} />
					</footer>
				</article>
			</li>
		{/each}
	</ul>

	{#if data.articles.length > 0}
		<Navigator page={data.page} has_more={data.has_more} />
	{/if}
</section>
