<script lang="ts">
	import { createArticle, sendArticle } from '../../lib/api';
	import { getAPIKey } from '../../lib/storage';
	import type { UserProfile } from '@savetoink/shared';

	let { profile }: { profile: UserProfile | null } = $props();

	let url = $state('');
	let loading = $state(false);
	let status: '' | 'success' | 'error' = $state('');
	let errorMessage = $state('');
	let statusTimeout: ReturnType<typeof setTimeout>;
	let sendToDevice = $derived(profile?.auto_send || false);

	async function getCurrentTabUrl(): Promise<string | null> {
		try {
			const [tab] = await browser.tabs.query({ active: true, currentWindow: true });
			const url = tab?.url ?? null;

			if (!url) {
				return null;
			}

			// Must be http:// or https://
			if (!url.startsWith('http://') && !url.startsWith('https://')) {
				return null;
			}

			// Exclude localhost and local network addresses
			try {
				const urlObj = new URL(url);
				const hostname = urlObj.hostname.toLowerCase();

				// Exclude localhost variants
				if (
					hostname === 'localhost' ||
					hostname === '127.0.0.1' ||
					hostname === '::1' ||
					hostname.endsWith('.local')
				) {
					return null;
				}
			} catch {
				return null;
			}

			return url;
		} catch (error) {
			console.error('failed to get current tab URL:', error);
			return null;
		}
	}

	async function loadCurrentTabUrl() {
		const currentUrl = await getCurrentTabUrl();
		if (currentUrl) {
			url = currentUrl;
		}
	}

	loadCurrentTabUrl();

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

			const article = await createArticle(url, apiKey);

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

<h2>Add Article</h2>

<p>Save a new article to your reading list</p>

<form onsubmit={handleSubmit}>
	<input
		type="url"
		name="url"
		placeholder="https://example.com/article"
		bind:value={url}
		required
	/>
	{#if profile?.device_email}
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
