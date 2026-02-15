<script lang="ts">
	import { page } from '$app/state';
</script>

<h1>error</h1>

{#if page.error}
	<p>{page.error.message}</p>
{/if}

<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
<a href="/">go home</a>
