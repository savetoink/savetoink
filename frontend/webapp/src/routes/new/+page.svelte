<script lang="ts">
	import { goto } from '$app/navigation';
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import type { PageData, ActionData } from './$types';
	import type { UserProfile, ArticleResponse } from '@savetoink/shared';

	let { data, form }: { data: PageData; form: ActionData } = $props();
	let user = $derived(data?.user as UserProfile | undefined);
	let sendToDevice = $derived(user?.auto_send);

	let formElement = $state<HTMLFormElement | undefined>();
	let isSubmitting = $state(false);
	let error = $state<string | null>(null);
	let success = $state(false);
	let article = $state<ArticleResponse | null>(null);

	onMount(() => {
		if (data.incomingUrl && formElement) {
			formElement.requestSubmit();
		}
	});

	async function handleEnhance() {
		error = null;
		isSubmitting = true;
		return async ({
			update
		}: {
			update: (options?: { reset?: boolean; invalidateAll?: boolean }) => Promise<void>;
		}) => {
			await update();
			isSubmitting = false;
			if (form?.success) {
				success = true;
				article = form.article;
				setTimeout(() => goto(resolve('/articles')), 2000);
			} else if (form?.error) {
				error = form.error;
			}
		};
	}
</script>

<section>
	<hgroup>
		<h1>Add Article</h1>
		<p>Save a new article to your reading list</p>
	</hgroup>

	<!-- svelte-ignore a11y_autofocus -->
	<form bind:this={formElement} method="POST" action="?/new" use:enhance={handleEnhance}>
		<fieldset>
			<label>
				URL
				<input
					type="url"
					name="url"
					required
					placeholder="https://example.com/article"
					autocomplete="url"
					autofocus
					value={data.incomingUrl ?? ''}
					disabled={isSubmitting}
				/>
				<small>Enter the full URL of the article you want to save</small>
			</label>
			{#if data.user?.device_email != ''}
				<label>
					<input
						type="checkbox"
						name="sendToDevice"
						bind:checked={sendToDevice}
						disabled={isSubmitting}
					/>
					Send to device: <code>{user?.device_email}</code>
				</label>
			{/if}
		</fieldset>
		<button type="submit" disabled={isSubmitting}>
			{#if isSubmitting}
				Adding...
			{:else}
				Add
			{/if}
		</button>
		{#if isSubmitting}
			<progress aria-busy="true"></progress>
		{/if}
		{#if error}
			<p style="color: red">{error}</p>
		{/if}
	</form>
</section>

{#if success}
	<dialog open>
		<article>
			<header>
				<strong>Article Added!</strong>
			</header>
			<p>Your article has been saved</p>
			<footer>
				<progress aria-busy="true"></progress>
				<small>Redirecting to articles...</small>
			</footer>
		</article>
	</dialog>
{/if}
