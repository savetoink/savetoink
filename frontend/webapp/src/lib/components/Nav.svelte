<script lang="ts">
	import NavItem from './NavItem.svelte';

	let { loggedIn }: { loggedIn: boolean } = $props();
</script>

<nav>
	<ul>
		{#if loggedIn}
			<NavItem link="/new" text="New" />

			<NavItem link="/articles" text="Articles" />

			<NavItem link="/articles?favorites=true" text="Favorites" />
		{/if}
		<NavItem link="/account" text="Account" />
	</ul>
</nav>

<style>
	ul {
		margin-top: 2rem;
		margin-bottom: 1rem;
		display: flex;
		flex-grow: 1;
		gap: 1.2rem;
		justify-content: center;
	}
</style>
