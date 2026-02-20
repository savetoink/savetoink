<script lang="ts">
	import { PUBLIC_AUTH_BACKEND } from '$env/static/public';
	import Auth0Login from './Auth0Login.svelte';
	import SharedKeyLogin from './SharedKeyLogin.svelte';
	import { Auth0, SharedKey } from '$lib/consts';

	import type { PageProps } from './$types';

	let { form, data }: PageProps = $props();
	let msg = $derived(
		data?.kindle_email
			? 'Email address associated with your Kindle device'
			: 'Add your Amazon Kindle Email to enable sending articles to your device'
	);
</script>

<section>
	<h2>Login</h2>
	{#if PUBLIC_AUTH_BACKEND === Auth0}
		<Auth0Login {data} />
	{:else if PUBLIC_AUTH_BACKEND === SharedKey}
		<SharedKeyLogin {form} />
	{/if}
</section>

<section>
	<h2>Kindle delivery</h2>
	<form method="POST" action="?/updateProfile">
		<fieldset>
			<legend>{msg}</legend>
			<input
				type="email"
				name="kindleEmail"
				required
				autocomplete="email"
				value={data?.kindle_email}
				disabled={data?.kindle_email}
				placeholder="yourname@kindle.com"
			/>
		</fieldset>
		{#if !data?.kindle_email}
			<button type="submit">Enable Kindle delivey</button>
		{/if}
	</form>

	{#if data?.kindle_email}
		<form method="POST" action="?/deleteProfile">
			<button type="submit">Disable Kindle delivery</button>
		</form>
	{/if}
</section>
