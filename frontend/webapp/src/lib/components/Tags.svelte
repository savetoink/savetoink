<script lang="ts">
	import type { Article } from '@savetoink/shared';
	import { resolve } from '$app/paths';

	let {
		tags,
		clickable = false,
		onTagClick = () => {}
	}: { tags: Article['tags']; clickable?: boolean; onTagClick?: (tag: string) => void } = $props();

	function handleTagClick(tag: string) {
		onTagClick(tag);
	}
</script>

{#if tags && tags.length > 0}
	<ul class="tags">
		{#each tags as tag (tag)}
			<li>
				{#if clickable}
					<a
						href={resolve(`/articles?tag=${encodeURIComponent(tag)}`)}
						onclick={() => handleTagClick(tag)}
					>
						#{tag}
					</a>
				{:else}
					<ins>#{tag}</ins>
				{/if}
			</li>
		{/each}
	</ul>
{/if}

<style>
	.tags {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		padding: 0;
		margin: 0;
		list-style: none;
	}

	.tags li {
		display: inline-block;
		background-color: var(--pico-primary-background);
		color: var(--pico-color);
		border-radius: var(--pico-border-radius);
		padding: 0.25rem 0.5rem;
	}

	.tags a {
		text-decoration: none;
		color: inherit;
	}

	.tags a:hover {
		text-decoration: underline;
	}

	.tags ins {
		text-decoration: none;
	}
</style>
