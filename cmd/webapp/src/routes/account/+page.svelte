<script lang="ts">
	import { PUBLIC_AUTH_BACKEND } from '$env/static/public';
	import { Auth0, SharedKey } from '$lib/consts';
	import { checkLoggedIn } from '$lib/auth';
	import Auth0Login from './Auth0Login.svelte';
	import SharedKeyLogin from './SharedKeyLogin.svelte';
	import KindleDelivery from './KindleDelivery.svelte';

	import type { PageProps } from './$types';

	let { form, data }: PageProps = $props();
	const loggedIn = $derived(checkLoggedIn(data));
</script>

<section>
	<h2>Login</h2>
	{#if PUBLIC_AUTH_BACKEND === Auth0}
		<Auth0Login {data} />
	{:else if PUBLIC_AUTH_BACKEND === SharedKey}
		<SharedKeyLogin {form} {data} />
	{/if}
</section>

{#if loggedIn}
	<KindleDelivery {data} />
{/if}
