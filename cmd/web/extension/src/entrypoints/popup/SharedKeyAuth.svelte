<script lang="ts">
	import { getProfile } from '../../lib/api';
	import { ApiError } from '@savetoink/shared';
	import type { UserProfile } from '@savetoink/shared';

	export let apiKey = '';
	export let profile: UserProfile | null = null;
	export let onSave: (detail: { apiKey: string; profile?: UserProfile }) => void | Promise<void>;
	export let onLogout: () => void | Promise<void>;

	let localKey = apiKey;
	let saveStatus = '';

	$: localKey = apiKey;
	$: isLoggedIn = apiKey && profile;

	async function handleSave() {
		try {
			saveStatus = 'validating';

			const controller = new AbortController();
			const timeoutId = setTimeout(() => controller.abort(), 10000);

			const userProfile = await getProfile(localKey);

			clearTimeout(timeoutId);

			await onSave({ apiKey: localKey, profile: userProfile });
			saveStatus = 'saved';
			setTimeout(() => {
				saveStatus = '';
			}, 2000);
		} catch (error) {
			if (error instanceof ApiError && (error.status === 401 || error.status === 403)) {
				saveStatus = 'invalid';
			} else {
				saveStatus = 'error';
				console.error('failed to validate API key:', error);
			}
		}
	}

	async function handleLogout() {
		await onLogout();
	}
</script>

{#if !isLoggedIn}
	<form on:submit|preventDefault={handleSave}>
		<label for="api-key">API Key</label>
		<input
			id="api-key"
			type="password"
			bind:value={localKey}
			placeholder="Enter your shared API key"
			required
			minlength="1"
		/>

		<button type="submit">Login</button>
		{#if saveStatus === 'validating'}
			<ins>validating API key...</ins>
		{:else if saveStatus === 'invalid'}
			<p aria-invalid="true" class="error">invalid API key</p>
		{:else if saveStatus === 'error'}
			<p aria-invalid="true" class="error">failed to save API key</p>
		{/if}
	</form>
{:else}
	<button type="button" on:click={handleLogout}>Logout</button>
{/if}
