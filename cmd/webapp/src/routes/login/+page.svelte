<script lang="ts">
	import { PUBLIC_AUTH_BACKEND } from '$env/static/public';
	import Auth0Login from './Auth0Login.svelte';
	import SharedKeyLogin from './SharedKeyLogin.svelte';
	import type { PageProps } from './$types';

	let { form }: { form: PageProps['form'] } = $props();
</script>

{#if PUBLIC_AUTH_BACKEND === 'auth0'}
	<Auth0Login />
{:else}
	<SharedKeyLogin {form} />
{/if}
