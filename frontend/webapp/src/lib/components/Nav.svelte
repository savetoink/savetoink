<script lang="ts">
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
			<NavItem link="/" text="My List" />

			<NavItem link="/?favorite=true" text="Favorites" />

			<NavItem link="/new" text="New" />
		{/if}
		<NavItem link="/account" text="Account" />
	</ul>
</nav>
