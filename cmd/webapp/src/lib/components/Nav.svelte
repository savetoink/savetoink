<script lang="ts">
	import { page } from '$app/state';
	import NavItem from './NavItem.svelte';
</script>

<nav>
	<ul>
		{#if page.data?.total && page.data.total > 0}
			<NavItem link="/" text="My List ({page.data.total})" />
		{:else}
			<NavItem link="/" text="My List" />
		{/if}
		<NavItem link="/new" text="New" />
		<NavItem link="/account" text="Account" />
	</ul>
</nav>
