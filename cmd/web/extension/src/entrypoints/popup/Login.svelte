<script lang="ts">
	import SharedKeyAuth from './SharedKeyAuth.svelte';
	import Auth0Auth from './Auth0Auth.svelte';
	import { SharedKey, Auth0 } from '@savetoink/shared';
	import type { AuthBackendType, components } from '@savetoink/shared';

	export let apiKey = '';
	export let authBackend: AuthBackendType | null = null;
	export let userProfile: components['schemas']['UserProfile'] | null = null;
	export let onApiKeySave: (detail: {
		apiKey: string;
		profile?: components['schemas']['UserProfile'];
	}) => void | Promise<void>;
	export let onApiKeyLogout: () => void | Promise<void>;
</script>

{#if authBackend === SharedKey}
	<SharedKeyAuth {apiKey} profile={userProfile} onSave={onApiKeySave} onLogout={onApiKeyLogout} />
{:else if authBackend === Auth0}
	<Auth0Auth {apiKey} profile={userProfile} onSave={onApiKeySave} onLogout={onApiKeyLogout} />
{/if}
