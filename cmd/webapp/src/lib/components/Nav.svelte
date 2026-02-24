<script lang="ts">
	import { page } from '$app/state';
	import NavItem from './NavItem.svelte';

	let { loggedIn }: { loggedIn: boolean } = $props();
</script>

<nav>
	<ul>
		{#if loggedIn}
			{#if page.data?.total && page.data.total > 0}
				<NavItem link="/" text="My List ({page.data.total})" />
			{:else}
				<NavItem link="/" text="My List" />
			{/if}
			<NavItem link="/new" text="New" />
			<NavItem link="/account" text="Account" />
		{/if}
	</ul>
</nav>
