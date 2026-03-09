<script lang="ts">
	import NavItem from './NavItem.svelte';

	let { loggedIn }: { loggedIn: boolean } = $props();
</script>

<nav>
	<ul>
		{#if loggedIn}
			<NavItem link="/new" text="Add new article" />

			<NavItem link="/articles" text="All" />

			<NavItem link="/articles?favorite=true" text="Favorites" />
		{/if}
		<NavItem link="/account" text="My account" />
	</ul>
</nav>

<style>
	ul {
		gap: 1rem;
	}
</style>
