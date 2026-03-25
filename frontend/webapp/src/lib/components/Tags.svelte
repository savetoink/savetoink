<script lang="ts">
	import { resolve } from '$app/paths';
	import type { Article } from '@savetoink/shared';

	let { tags }: { tags: Article['tags'] } = $props();
</script>

{#if tags && tags.length > 0}
	<ul>
		{#each tags as tag (tag)}
			<li>
				<a href={resolve(`/articles?tag=${tag}`)}>
					<ins>#{tag}</ins>
				</a>
			</li>
		{/each}
	</ul>
{/if}

<style>
	ul {
		padding: 0;
		display: flex;
		gap: 0.5rem;
	}

	li {
		list-style: none;
	}

	a {
		text-decoration: none;
	}
</style>
