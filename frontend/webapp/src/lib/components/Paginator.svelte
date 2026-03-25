<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';

	let {
		page,
		hasMore: has_more,
		tag,
		favorite
	}: {
		page: number;
		hasMore: boolean;
		tag?: string;
		favorite?: boolean;
	} = $props();

	function navigateTo(newPage: number) {
		const params = new SvelteURLSearchParams();
		params.set('page', String(newPage));
		if (tag) {
			params.set('tag', tag);
		}
		if (favorite) {
			params.set('favorite', 'true');
		}
		goto(resolve(`/articles?${params.toString()}` as unknown as '/'));
	}
</script>

<nav aria-label="Pagination">
	{#if page > 1}
		<button onclick={() => navigateTo(page - 1)}>Previous</button>
	{/if}
	{#if has_more}
		<button onclick={() => navigateTo(page + 1)}>Next</button>
	{/if}
</nav>

<style>
	nav {
		display: flex;
		justify-content: center;
		gap: 1rem;
	}
</style>
