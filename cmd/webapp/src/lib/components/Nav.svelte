<script lang="ts">
	import { page } from '$app/state';
	import NavItem from './NavItem.svelte';

	let { loggedIn }: { loggedIn: boolean } = $props();
</script>

<nav>
	<ul>
		{#if loggedIn}
			<strong>
				<li>Save<span>to.</span>ink</li>
			</strong>
			{#if page.data?.total && page.data.total > 0 && !page.url.search.includes('favorite=true')}
				<NavItem link="/" text="My List ({page.data.total})" />
			{:else}
				<NavItem link="/" text="My List" />
			{/if}

			{#if page.data?.total && page.data.total > 0 && page.url.search.includes('favorite=true')}
				<NavItem link="/?favorite=true" text="Favorites ({page.data.total})" />
			{:else}
				<NavItem link="/?favorite=true" text="Favorites" />
			{/if}

			<NavItem link="/new" text="New" />
		{/if}
		<NavItem link="/account" text="Account" />
	</ul>
</nav>

<style>
	span {
		color: var(--pico-primary);
	}
</style>
