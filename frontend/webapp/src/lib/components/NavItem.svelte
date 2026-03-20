<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	let { link, text } = $props();

	const linkUrl = $derived(
		new URL(link, typeof window !== 'undefined' ? window.location.origin : 'http://example.com')
	);
</script>

<li>
	<a
		href={resolve(link)}
		aria-current={page.url.pathname === linkUrl.pathname && page.url.search === linkUrl.search
			? 'page'
			: undefined}><button>{text}</button></a
	>
</li>

<style>
	button {
		font-size: 1.5rem;
	}
</style>
