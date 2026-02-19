<script lang="ts">
	import { PUBLIC_AUTH_BACKEND } from '$env/static/public';
	import Auth0Login from './Auth0Login.svelte';
	import SharedKeyLogin from './SharedKeyLogin.svelte';
	import { Auth0, SharedKey } from '$lib/consts';

	import type { PageProps } from './$types';

	let { form }: PageProps = $props();
</script>

<section>
	{#if PUBLIC_AUTH_BACKEND === Auth0}
		<Auth0Login />
	{:else if PUBLIC_AUTH_BACKEND === SharedKey}
		<SharedKeyLogin {form} />
	{/if}
</section>

<section>
	<form method="POST" action="?/profile">
		<label>
			Kindle Email
			<input type="email" name="kindleEmail" required autocomplete="email" />
		</label>
		<button type="submit">Save</button>
	</form>
</section>
