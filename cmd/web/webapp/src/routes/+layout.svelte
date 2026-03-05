<script lang="ts">
	import { navigating } from '$app/state';
	import Nav from '$lib/components/Nav.svelte';
	import Landing from '$lib/components/Landing.svelte';
	import { isDev } from '$lib/consts';
	import '@savetoink/shared/css';

	if (isDev) {
		import('@savetoink/shared/css-dev');
	}

	const version = __APP_VERSION__;
	const buildDate = __BUILD_DATE__;
	const gitHash = __GIT_HASH__;
	const versionTxt = isDev ? `${version}-${buildDate}-${gitHash}` : version;
	const title = isDev ? `Save to Ink - ${versionTxt}` : 'Save to Ink';

	let { children, data } = $props();
	const loggedIn = $derived(data.isLoggedIn);
</script>

<svelte:head>
	<link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png" />
	<link rel="apple-touch-icon" sizes="152x152" href="/apple-touch-icon-152x152.png" />
	<link rel="apple-touch-icon" sizes="144x144" href="/apple-touch-icon-144x144.png" />
	<link rel="apple-touch-icon" sizes="120x120" href="/apple-touch-icon-120x120.png" />
	<link rel="apple-touch-icon" sizes="76x76" href="/apple-touch-icon-76x76.png" />
	<link rel="apple-touch-icon" sizes="60x60" href="/apple-touch-icon-60x60.png" />
	<link rel="icon" type="image/png" sizes="192x192" href="/android-icon-192x192.png" />
	<link rel="icon" type="image/png" sizes="512x512" href="/android-icon-512x512.png" />
	<link rel="icon" href="/favicon.ico" />
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
