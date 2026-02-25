<script>
	import { resolve } from '$app/paths';
	import ArticleControls from './ArticleControls.svelte';

	let { article } = $props();
</script>

<li>
	<article>
		<header>
			{#if article.title}
				<h2>
					<a href={resolve(`/articles/${article.id}`)}
						>{#if article.favorite}
							<span>⭐️&nbsp;</span>
						{/if}
						{article.title}</a
					>
				</h2>
			{:else}
				<h2>
					<a href={resolve(`/articles/${article.id}`)}
						>{#if article.favorite}
							<span>⭐️&nbsp;</span>
						{/if}
						{article.url}</a
					>
				</h2>
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

<style>
	li {
		list-style: none;
	}

	header {
		margin: 0;
	}

	footer {
		margin: 0;
	}
</style>
