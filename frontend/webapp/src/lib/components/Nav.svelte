<script lang="ts">
	import { page } from '$app/state';
	import NavItem from './NavItem.svelte';
	import { Landing } from '@savetoink/shared/components';

	let { loggedIn }: { loggedIn: boolean } = $props();
</script>

<nav>
	<ul>
		<li>
			<strong>
				<Landing />
			</strong>
		</li>
		{#if loggedIn}
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
