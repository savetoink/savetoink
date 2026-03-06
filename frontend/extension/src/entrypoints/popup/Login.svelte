<script lang="ts">
	import SharedKeyAuth from './SharedKeyAuth.svelte';
	import Auth0Auth from './Auth0Auth.svelte';
	import { SharedKey, Auth0 } from '@savetoink/shared';
	import type { AuthBackendType, UserProfile } from '@savetoink/shared';

	let {
		apiKey,
		authBackend,
		userProfile,
		onApiKeySave,
		onApiKeyLogout
	}: {
		apiKey: string;
		authBackend: AuthBackendType | null;
		userProfile: UserProfile | null;
		onApiKeySave: (detail: { apiKey: string; profile?: UserProfile }) => void | Promise<void>;
		onApiKeyLogout: () => void | Promise<void>;
	} = $props();
</script>

{#if authBackend === SharedKey}
	<SharedKeyAuth {apiKey} profile={userProfile} onSave={onApiKeySave} onLogout={onApiKeyLogout} />
{:else if authBackend === Auth0}
	<Auth0Auth {apiKey} profile={userProfile} onSave={onApiKeySave} onLogout={onApiKeyLogout} />
{/if}
