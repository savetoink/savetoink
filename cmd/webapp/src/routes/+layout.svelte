<script lang="ts">
	import { navigating } from '$app/state';
	import favicon from '$lib/assets/favicon.svg';
	import Nav from '$lib/components/Nav.svelte';
	import Landing from '$lib/components/Landing.svelte';
	import { isDev } from '$lib/consts';

	const version = __APP_VERSION__;
	const buildDate = __BUILD_DATE__;
	const gitHash = __GIT_HASH__;
	const versionTxt = isDev ? `${version}-${buildDate}-${gitHash}` : version;
	const title = 'Save to Ink';

	let { children, data } = $props();
	const loggedIn = $derived(data.isLoggedIn);
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>{title}</title>
</svelte:head>

<header class="container">
	<Nav {loggedIn} />
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
	<small>
		<Landing />
		- {versionTxt} - Source on
		<a
			href="https://github.com/savetoink/savetoink"
			target="_blank"
			rel="external noopener noreferrer">GitHub</a
		></small
	>
</footer>

<style>
	footer {
		text-align: center;
		margin-bottom: 1rem;
	}
</style>
