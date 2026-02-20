<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';

	let { page, has_more }: { page: number; has_more: boolean } = $props();

	function navigateTo(newPage: number) {
		goto(resolve(`/?page=${newPage}` as unknown as '/'));
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
