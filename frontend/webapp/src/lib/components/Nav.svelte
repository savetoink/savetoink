<script lang="ts">
	import NavItem from './NavItem.svelte';

	let { loggedIn }: { loggedIn: boolean } = $props();
</script>

<nav>
	<ul>
		{#if loggedIn}
			<NavItem link="/" text="My List" />

			<NavItem link="/?favorite=true" text="Favorites" />

			<NavItem link="/new" text="New" />
		{/if}
		<NavItem link="/account" text="Account" />
	</ul>
</nav>
