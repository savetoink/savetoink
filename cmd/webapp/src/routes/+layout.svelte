<script lang="ts">
	import { navigating } from '$app/state';
	import favicon from '$lib/assets/favicon.svg';
	import Nav from '$lib/components/Nav.svelte';
	let { children } = $props();
	const version = __APP_VERSION__;
	const buildDate = __BUILD_DATE__;
	const gitHash = __GIT_HASH__;
	const title = `Save to Ink - ${version}-${buildDate}-${gitHash}`;
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>{title}</title>
</svelte:head>

<Nav />

<main>
	{#if navigating.to}
		<p>loading...</p>
	{:else}
		{@render children()}
	{/if}
</main>

<hr />

<footer>
	<p>{title}</p>
</footer>
