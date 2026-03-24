<script lang="ts">
	import { goto } from '$app/navigation';
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import type { PageData, ActionData } from './$types';
	import type { UserProfile } from '@savetoink/shared';

	const MAX_TAGS = 10;
	const MAX_TAG_LENGTH = 50;

	let { data, form }: { data: PageData; form: ActionData } = $props();
	let user = $derived(data?.user as UserProfile | undefined);
	let sendToDevice = $derived(user?.auto_send);

	let formElement = $state<HTMLFormElement | undefined>();
	let isSubmitting = $state(false);
	let error = $state<string | null>(null);
	let success = $state(false);
	let tagsInput = $state('');
	let tagsError = $state<string | null>(null);

	let parsedTags = $derived(() => {
		if (!tagsInput) return [];
		return tagsInput
			.split(',')
			.map((tag) => tag.trim())
			.filter((tag) => tag.length > 0);
	});

	function validateTags(tags: string[]): string | null {
		if (tags.length > MAX_TAGS) {
			return `Maximum ${MAX_TAGS} tags allowed per article`;
		}
		for (const tag of tags) {
			if (tag.length > MAX_TAG_LENGTH) {
				return `Tag "${tag}" exceeds maximum length of ${MAX_TAG_LENGTH} characters`;
			}
		}
		return null;
	}

	onMount(() => {
		if (data.incomingUrl && formElement) {
			formElement.requestSubmit();
		}
	});

	async function handleEnhance() {
		error = null;
		tagsError = null;

		// Validate tags before submission
		const tags = parsedTags();
		const validationError = validateTags(tags);
		if (validationError) {
			tagsError = validationError;
			return;
		}

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
			<label>
				Tags
				<input
					type="text"
					name="tags"
					placeholder="reading, tech, tutorial"
					bind:value={tagsInput}
					disabled={isSubmitting}
				/>
				<small>
					Comma-separated list tags (max {MAX_TAGS} tags, {MAX_TAG_LENGTH} characters each).
					{(parsedTags().length > 0 && ` Current: ${parsedTags().length}/${MAX_TAGS}`) || undefined}
				</small>
			</label>
			{#if tagsError}
				<p style="color: red" role="alert">{tagsError}</p>
			{/if}
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
			<p style="color: red" role="alert">{error}</p>
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
