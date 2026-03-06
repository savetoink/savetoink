<script lang="ts">
	import { getProfile } from '../../lib/api';
	import { ApiError } from '@savetoink/shared';
	import type { UserProfile } from '@savetoink/shared';

	let {
		apiKey,
		profile,
		onSave,
		onLogout
	}: {
		apiKey: string;
		profile: UserProfile | null;
		onSave: (detail: { apiKey: string; profile?: UserProfile }) => void | Promise<void>;
		onLogout: () => void | Promise<void>;
	} = $props();

	let localKey = $derived(apiKey);
	let isLoggedIn = $derived(!!apiKey && !!profile);
	let saveStatus = $state('');

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
	<form
		onsubmit={(e) => {
			e.preventDefault();
			handleSave();
		}}
	>
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
	<button type="button" onclick={handleLogout}>Logout</button>
{/if}
