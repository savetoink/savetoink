<script lang="ts">
	import Login from './Login.svelte';
	import { saveAPIKey, clearAPIKey } from '../../lib/storage';
	import { API_URL, isDev } from '../../lib/api';
	import { SharedKey } from '@savetoink/shared';
	import type { UserProfile, AuthBackendType } from '@savetoink/shared';

	let {
		profile = $bindable(),
		apiKey = $bindable()
	}: {
		profile: UserProfile | null;
		apiKey: string;
	} = $props();

	let authBackend: AuthBackendType =
		(import.meta.env.PUBLIC_AUTH_BACKEND as AuthBackendType) || SharedKey;

	async function handleApiKeySave(detail: { apiKey: string; profile?: UserProfile }) {
		await saveAPIKey(detail.apiKey);
		apiKey = detail.apiKey;
		if (detail.profile) {
			profile = detail.profile;
		}
	}

	async function handleApiKeyLogout() {
		await clearAPIKey();
		apiKey = '';
		profile = null;
	}
</script>

<h2>Account</h2>

<section>
	<Login
		{apiKey}
		{authBackend}
		userProfile={profile}
		onApiKeySave={handleApiKeySave}
		onApiKeyLogout={handleApiKeyLogout}
	/>
</section>

<section>
	<ul>
		{#if profile}
			{#if profile.email && authBackend !== SharedKey}
				<li>
					<strong>Email:</strong>
					{profile.email}
				</li>
			{/if}
			<li>
				<strong>Device email:</strong>
				{profile.device_email || 'Not set'}
			</li>
			<li>
				<strong>Auto-send:</strong>
				{profile.auto_send ? 'Enabled' : 'Disabled'}
			</li>
		{/if}
		{#if isDev}
			<li>
				<strong>API URL:</strong>
				{API_URL}
			</li>
		{/if}
	</ul>
</section>

<style>
	ul li {
		list-style: none;
	}
</style>
