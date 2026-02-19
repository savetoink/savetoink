<script lang="ts">
	import { PUBLIC_AUTH_BACKEND } from '$env/static/public';
	import Auth0Login from './Auth0Login.svelte';
	import SharedKeyLogin from './SharedKeyLogin.svelte';
	import { Auth0, SharedKey } from '$lib/consts';

	import type { PageProps } from './$types';

	let { form, data }: PageProps = $props();
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
			Save your Amazon Kindle Email to enable sending articles to your device
			<br />

			<input
				type="email"
				name="kindleEmail"
				required
				autocomplete="email"
				value={data?.kindle_email}
			/>
		</label>
		<button type="submit">Save</button>
	</form>
</section>
