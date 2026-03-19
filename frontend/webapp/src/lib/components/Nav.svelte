<script lang="ts">
	import NavItem from './NavItem.svelte';

	let { loggedIn }: { loggedIn: boolean } = $props();
</script>

<nav>
	<ul>
		{#if loggedIn}
			<NavItem link="/new" text="New" />

			<NavItem link="/articles" text="Articles" />
		{/if}
		<NavItem link="/account" text="Account" />
	</ul>
</nav>

<style>
	ul {
		gap: 1rem;
	}
</style>
