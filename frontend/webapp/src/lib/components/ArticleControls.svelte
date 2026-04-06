<script lang="ts">
	import type { Article, UserProfile } from '@savetoink/shared';
	import { deleteArticle as deleteArticleAction, MAX_TAGS } from '@savetoink/shared';

	let {
		article,
		user,
		favoriteForm,
		sendForm,
		deleteForm,
		favoriteSubmitting = false,
		sendSubmitting = false,
		deleteSubmitting = false,
		onAddTag,
		tagCount = 0,
		maxTags = MAX_TAGS,
		showAddTagForm = false
	}: {
		article: Article;
		user?: UserProfile;
		favoriteForm?: HTMLFormElement;
		sendForm?: HTMLFormElement;
		deleteForm?: HTMLFormElement;
		favoriteSubmitting?: boolean;
		sendSubmitting?: boolean;
		deleteSubmitting?: boolean;
		onAddTag?: () => void;
		tagCount?: number;
		maxTags?: number;
		showAddTagForm?: boolean;
	} = $props();

	const canSendToDevice = $derived(!!user?.device_email);
	const atMaxTags = $derived(tagCount >= maxTags);

	async function handleFavoriteClick() {
		favoriteForm?.requestSubmit();
	}

	async function handleSendClick() {
		sendForm?.requestSubmit();
	}

	async function handleDeleteClick() {
		if (!deleteForm) return;
		deleteArticleAction(deleteForm);
	}

	function handleAddTagClick() {
		onAddTag?.();
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

	{#if onAddTag}
		<button type="button" onclick={handleAddTagClick} disabled={atMaxTags && !showAddTagForm}
			>{showAddTagForm ? 'Close' : 'Add tag'}</button
		>
	{/if}
</div>

<style>
	div {
		display: flex;
		gap: 1rem;
		justify-content: center;
	}
</style>
