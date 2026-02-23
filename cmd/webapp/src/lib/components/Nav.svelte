<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/stores';
</script>

<nav>
	<ul>
		{#if $page.data?.total && $page.data.total > 0}
			<li><a href={resolve('/')}>My List ({$page.data.total})</a></li>
		{:else}
			<li><a href={resolve('/')}>My List</a></li>
		{/if}
		<li><a href={resolve('/new')}>New</a></li>
		<li><a href={resolve('/account')}>Account</a></li>
	</ul>
</nav>
