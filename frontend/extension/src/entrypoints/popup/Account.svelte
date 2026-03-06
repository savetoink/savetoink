<script lang="ts">
	import Login from './Login.svelte';
	import { saveAPIKey, clearAPIKey } from '../../lib/storage';
	import { API_URL } from '../../lib/api';
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

<h1>Account</h1>

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
			{#if profile.email}
				<li>
					<strong>Email:</strong>
					{profile.email}
				</li>
			{/if}

			{#if profile.device_email}
				<li>
					<strong>Device email:</strong>
					{profile.device_email || 'Not set'}
				</li>
			{/if}

			{#if profile.auto_send}
				<li>
					<strong>Auto-send:</strong>
					{profile.auto_send ? 'Enabled' : 'Disabled'}
				</li>
			{/if}
		{/if}
		{#if import.meta.env.DEV}
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
