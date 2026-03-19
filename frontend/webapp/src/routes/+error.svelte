<script lang="ts">
	import { page } from '$app/stores';
	import { resolve } from '$app/paths';
</script>

{#if $page.status >= 500}
	<section>
		<hgroup>
			<h1>Server Error</h1>
			<p>Something went wrong on our end. Please try again later.</p>
		</hgroup>
		{#if $page.error?.message}
			<details>
				<summary>Error Details</summary>
				<pre>{$page.error.message}</pre>
			</details>
		{/if}
		<nav>
			<a href={resolve('/')}>Go Home</a>
		</nav>
	</section>
{:else if $page.status === 404}
	<section>
		<hgroup>
			<h1>Not Found</h1>
			<p>The page you're looking for doesn't exist.</p>
		</hgroup>
	</section>
{:else}
	<section>
		<hgroup>
			<h1>Error</h1>
			<p>An unexpected error occurred.</p>
		</hgroup>
		{#if $page.error?.message}
			<details>
				<summary>Error Details</summary>
				<pre>{$page.error.message}</pre>
			</details>
		{/if}
	</section>
{/if}
