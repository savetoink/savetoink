<script lang="ts">
	import { resolve } from '$app/paths';
	import type { Article } from '@savetoink/shared';

	let {
		article,
		selected = false,
		ref
	}: {
		article: Article;
		selected: boolean;
		ref?: (el: HTMLLIElement) => void;
	} = $props();
	let liElement: HTMLLIElement | undefined;

	let title: string = $derived(article.title ?? article.url);

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
						{#if article.author}
							{article.author} -
						{/if}
						{title}
					</a>
				</h2>
			{/if}
			{#if article.excerpt}
				<p>{article.excerpt}</p>
			{/if}
		</header>
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

		a {
			text-decoration: none;
		}
	}

	footer {
		margin: 0;
	}
</style>
