<script lang="ts">
	import { enhance } from '$app/forms';
	import { DeviceDomains } from '$lib/consts';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const deviceEmail = $derived(data?.user?.deviceEmail);
	let autoSend = $derived(data?.user?.autoSend || false);
	let profileForm = $state<HTMLFormElement | undefined>();
	let isSaving = $state(false);
	let error = $state<string | null>(null);

	async function handleEnhance() {
		isSaving = true;
		error = null;
		return async ({ result }: { result: { type: string } }) => {
			isSaving = false;
			if (result.type === 'failure') {
				error = 'Failed to save preference';
			}
		};
	}

	function handleAutoSendChange() {
		if (profileForm) {
			profileForm.requestSubmit();
		}
	}
</script>

<section>
	<h2>Device delivery (Kindle, Kobo, Tolino, etc.)</h2>
	{#if deviceEmail}
		<p>Email delivery enabled for <code>{deviceEmail}</code></p>
		<form
			bind:this={profileForm}
			method="POST"
			action="?/updateProfile"
			use:enhance={handleEnhance}
		>
			<label>
				<input
					type="checkbox"
					name="autoSend"
					bind:checked={autoSend}
					disabled={isSaving}
					onchange={handleAutoSendChange}
				/>
				Automatically send new articles to device
			</label>
			{#if isSaving}
				<small>Saving...</small>
			{/if}
			{#if error}
				<small style="color: red">{error}</small>
			{/if}
		</form>
		<form method="POST" action="?/deleteProfile">
			<button type="submit">Disable device delivery</button>
		</form>
	{:else}
		<form method="POST" action="?/updateProfile">
			<fieldset>
				<legend>
					<ol>
						<li>
							Add <code>no-reply@saveto.ink</code> to your Kindle's
							<a
								href="https://www.amazon.com/gp/sendtokindle/email/"
								target="_blank"
								rel="external noopener">Approved Personal Document E-mail List</a
							>
							or similar device delivery service
						</li>
						<li>
							<p>Save your device's email address to enable delivery</p>
						</li>
					</ol>
					<p></p>
				</legend>

				<input
					type="email"
					name="deviceEmail"
					required
					autocomplete="email"
					placeholder={DeviceDomains.join(', ')}
				/>
				<label>
					<input type="checkbox" name="autoSend" bind:checked={autoSend} />
					Automatically send new articles to device
				</label>
			</fieldset>
			<button type="submit">Enable delivery</button>
		</form>
	{/if}
</section>
