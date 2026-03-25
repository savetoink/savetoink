<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';

	// eslint-disable-next-line svelte/no-unused-props
	let { data }: { data: { tags: { tag: string; count: number }[]; user?: unknown } } = $props();

	function filterByTag(tag: string) {
		goto(resolve(`/articles?tag=${encodeURIComponent(tag)}`));
	}
</script>

<section>
	<hgroup>
		<h1>Tags</h1>
		<p>Browse and manage your article tags</p>
	</hgroup>

	{#if data.tags.length === 0}
		<p>No tags yet. Add tags to your articles to see them here.</p>
	{:else}
		<ul class="tags-cloud">
			{#each data.tags as { tag, count } (tag)}
				<li>
					<button type="button" onclick={() => filterByTag(tag)}>
						<span class="tag-name">#{tag}</span>
						<span class="count">({count})</span>
					</button>
				</li>
			{/each}
		</ul>
	{/if}
</section>

<style>
	hgroup {
		text-align: center;
		margin-bottom: 2rem;
	}

	hgroup h1 {
		margin: 0;
	}

	hgroup p {
		color: var(--muted-color);
		margin: 0.5rem 0 0 0;
	}

	.tags-cloud {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		padding: 0;
		justify-content: center;
		margin: 0;
	}

	.tags-cloud li {
		list-style: none;
	}

	.tags-cloud button {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		padding: 0.5rem 0.75rem;
		background-color: var(--pico-primary-background);
		color: var(--pico-color);
		border: none;
		border-radius: var(--pico-border-radius);
		cursor: pointer;
		transition:
			background-color 0.2s,
			transform 0.1s;
	}

	.tags-cloud button:hover {
		background-color: var(--pico-primary);
		transform: translateY(-1px);
	}

	.tags-cloud button:active {
		transform: translateY(0);
	}

	.tag-name {
		font-weight: 500;
	}

	.count {
		font-size: 0.8em;
		opacity: 0.7;
	}
</style>
