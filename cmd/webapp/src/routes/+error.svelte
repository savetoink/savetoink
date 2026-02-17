<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
</script>

<h1>error</h1>

{#if page.error}
	<p>{page.error.message}</p>
{/if}

<a href={resolve('/')}>go home</a>
