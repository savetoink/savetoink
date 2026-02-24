<script lang="ts">
	import { resolve } from '$app/paths';
	import ArticleControls from '$lib/components/ArticleControls.svelte';
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
	{/if}

	{#if data.articles.length > 0}
		<Navigator page={data.page} has_more={data.has_more} />
	{/if}
</section>

<style>
	ul {
		padding-left: 0;

		li {
			list-style: none;
		}
	}
</style>
