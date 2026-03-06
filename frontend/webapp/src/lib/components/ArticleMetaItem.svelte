<script lang="ts">
	import { resolve } from '$app/paths';
	import ArticleControls from './ArticleControls.svelte';

	let { article, user, selected = false, ref } = $props();
	let liElement: HTMLLIElement | undefined;

	$effect(() => {
		if (ref && liElement) {
			ref(liElement);
		}
	});
</script>

<li bind:this={liElement} data-selected={selected}>
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
			<ArticleControls {article} {user} />
		</footer>
	</article>
</li>

<style>
	li {
		list-style: none;
	}

	li[data-selected='true'] {
		outline: 2px solid var(--pico-primary-focus);
		outline-offset: -2px;
		background-color: var(--pico-card-sectioning-background-color);
		box-shadow: var(--pico-box-shadow);
		border-radius: var(--pico-border-radius);
	}

	header {
		margin: 0;
	}

	footer {
		margin: 0;
	}
</style>
