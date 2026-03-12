<script lang="ts">
	import { enhance } from '$app/forms';
	import type { Article, UserProfile } from '@savetoink/shared';

	let {
		article,
		user,
		favoriteSubmitting = false,
		sendSubmitting = false,
		deleteSubmitting = false
	}: {
		article: Article;
		user?: UserProfile;
		favoriteSubmitting?: boolean;
		sendSubmitting?: boolean;
		deleteSubmitting?: boolean;
	} = $props();

	const canSendToDevice = $derived(!!user?.device_email);

	async function handleDeleteEnhance({ cancel }: { cancel: () => void }) {
		if (!window.confirm('Are you sure you want to delete this article?')) {
			cancel();
		}
	}
</script>

<div>
	<form method="POST" action="/articles/{article.id}?/favorite" use:enhance>
		<button type="submit" aria-busy={favoriteSubmitting}
			>{article.favorite ? 'Unfavorite' : 'Favorite'}</button
		>
	</form>

	{#if canSendToDevice}
		<form method="POST" action="/articles/{article.id}?/send" use:enhance>
			<button type="submit" aria-busy={sendSubmitting}>Send</button>
		</form>
	{/if}

	<form method="POST" action="/articles/{article.id}?/delete" use:enhance={handleDeleteEnhance}>
		<button type="submit" aria-busy={deleteSubmitting}>Delete</button>
	</form>
</div>

<style>
	div {
		display: flex;
		gap: 1rem;
		justify-content: center;
	}
</style>
