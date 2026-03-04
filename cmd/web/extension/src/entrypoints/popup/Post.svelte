<script lang="ts">
	import { createArticle, sendArticle } from '../../lib/api';
	import { getAPIKey, getUserProfile } from '../../lib/storage';
	import type { UserProfile } from '@savetoink/shared';

	let url = $state('');
	let loading = $state(false);
	let status: '' | 'success' | 'error' = $state('');
	let errorMessage = $state('');
	let statusTimeout: ReturnType<typeof setTimeout>;
	let profile: UserProfile | null = $state(null);
	let sendToDevice = $state(false);

	async function getCurrentTabUrl(): Promise<string | null> {
		try {
			const [tab] = await browser.tabs.query({ active: true, currentWindow: true });
			return tab?.url ?? null;
		} catch (error) {
			console.error('failed to get current tab URL:', error);
			return null;
		}
	}

	function isValidUrl(url: string): boolean {
		try {
			const parsed = new URL(url);
			return parsed.protocol === 'http:' || parsed.protocol === 'https:';
		} catch {
			return false;
		}
	}

	async function loadProfile() {
		const apiKey = await getAPIKey();
		if (apiKey) {
			profile = await getUserProfile();
			sendToDevice = profile?.autoSend || false;
		}

		const currentUrl = await getCurrentTabUrl();
		if (currentUrl && isValidUrl(currentUrl)) {
			url = currentUrl;
		}
	}

	loadProfile();

	function clearStatus() {
		status = '';
		errorMessage = '';
	}

	async function handleSubmit(event: Event) {
		event.preventDefault();

		if (!url) {
			return;
		}

		clearTimeout(statusTimeout);
		loading = true;
		status = '';
		errorMessage = '';

		try {
			const apiKey = await getAPIKey();
			if (!apiKey) {
				errorMessage = 'please login first';
				status = 'error';
				loading = false;
				statusTimeout = setTimeout(clearStatus, 5000);
				return;
			}

			const article = await createArticle(url, null, apiKey);

			if (sendToDevice) {
				await sendArticle(article.id, apiKey);
			}

			status = 'success';

			statusTimeout = setTimeout(() => {
				window.close();
			}, 2000);
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'failed to create article';
			status = 'error';
			loading = false;
			statusTimeout = setTimeout(clearStatus, 5000);
		} finally {
			if (status !== 'error') {
				loading = false;
			}
		}
	}
</script>

<h1>Add Article</h1>

<p>Save a new article to your reading list</p>

<form onsubmit={handleSubmit}>
	<input
		type="url"
		name="url"
		placeholder="https://example.com/article"
		bind:value={url}
		required
	/>
	{#if profile?.deviceEmail}
		<label>
			<input type="checkbox" bind:checked={sendToDevice} />
			Send to reader device
		</label>
	{/if}
	<button type="submit" disabled={loading}>
		{loading ? 'Adding...' : 'Add'}
	</button>

	{#if status === 'error'}
		<p aria-invalid="true" class="error">
			{errorMessage}
		</p>
	{:else if status === 'success'}
		{#if sendToDevice}
			<ins>article saved and delivered successfully</ins>
		{:else}
			<ins>article saved successfully</ins>
		{/if}
	{/if}
</form>
