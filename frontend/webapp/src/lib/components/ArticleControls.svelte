<script lang="ts">
	import type { Article, UserProfile } from '@savetoink/shared';

	let {
		article,
		user,
		favoriteForm,
		sendForm,
		deleteForm,
		favoriteSubmitting = false,
		sendSubmitting = false,
		deleteSubmitting = false
	}: {
		article: Article;
		user?: UserProfile;
		favoriteForm?: HTMLFormElement;
		sendForm?: HTMLFormElement;
		deleteForm?: HTMLFormElement;
		favoriteSubmitting?: boolean;
		sendSubmitting?: boolean;
		deleteSubmitting?: boolean;
	} = $props();

	const canSendToDevice = $derived(!!user?.device_email);

	async function handleFavoriteClick() {
		favoriteForm?.requestSubmit();
	}

	async function handleSendClick() {
		sendForm?.requestSubmit();
	}

	async function handleDeleteClick() {
		if (!window.confirm('Are you sure you want to delete this article?')) {
			return;
		}
		deleteForm?.requestSubmit();
	}
</script>

<div>
	<button type="button" aria-busy={favoriteSubmitting} onclick={handleFavoriteClick}
		>{article.favorite ? 'Unfavorite' : 'Favorite'}</button
	>

	{#if canSendToDevice}
		<button type="button" aria-busy={sendSubmitting} onclick={handleSendClick}>Send</button>
	{/if}

	<button type="button" aria-busy={deleteSubmitting} onclick={handleDeleteClick}>Delete</button>
</div>

<style>
	div {
		display: flex;
		gap: 1rem;
		justify-content: center;
	}
</style>
