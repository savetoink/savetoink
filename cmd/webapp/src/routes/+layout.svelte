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

<header class="container">
	<Nav />
</header>

<main class="container">
	{#if navigating.to}
		<progress></progress>
	{:else}
		{@render children()}
	{/if}
</main>

<footer class="container">
	<hr />
	<small>{title} - Source on <a href="https://github.com/savetoink/savetoink">GitHub</a></small>
</footer>

<style>
	footer {
		text-align: center;
		margin-bottom: 1rem;
	}
</style>
