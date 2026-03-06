<script lang="ts">
	import { PUBLIC_AUTH_BACKEND } from '$env/static/public';
	import { Auth0, SharedKey } from '@savetoink/shared';
	import Auth0Login from './Auth0Login.svelte';
	import SharedKeyLogin from './SharedKeyLogin.svelte';
	import DeviceDelivery from './DeviceDelivery.svelte';

	import type { PageProps } from './$types';

	let { form, data }: PageProps = $props();
</script>

<section>
	<h2>Your account</h2>
	{#if PUBLIC_AUTH_BACKEND === Auth0}
		<Auth0Login {data} />
	{:else if PUBLIC_AUTH_BACKEND === SharedKey}
		<SharedKeyLogin {form} {data} />
	{/if}
</section>

{#if data?.isLoggedIn}
	<DeviceDelivery {data} />
{/if}
