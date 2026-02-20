<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
</script>

<article>
	<h1>Error</h1>
	{#if page.error}
		<p role="alert">{page.error.message}</p>
	{/if}
	<nav>
		<a href={resolve('/')}>Go home</a>
	</nav>
</article>
