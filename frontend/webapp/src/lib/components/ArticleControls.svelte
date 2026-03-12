<script lang="ts">
	import { enhance } from '$app/forms';
	import type { Article, UserProfile } from '@savetoink/shared';

	let { article, user }: { article: Article; user?: UserProfile } = $props();

	let favoriteSubmitting = $state(false);
	let sendSubmitting = $state(false);
	let deleting = $state(false);

	const canSendToDevice = $derived(!!user?.device_email);

	async function handleDeleteEnhance({ cancel }: { cancel: () => void }) {
		if (!window.confirm('Are you sure you want to delete this article?')) {
			cancel();
			return;
		}
		deleting = true;
		return async ({
			update
		}: {
			update: (options?: { reset?: boolean; invalidateAll?: boolean }) => Promise<void>;
		}) => {
			await update();
			deleting = false;
		};
	}

	async function handleFavoriteEnhance() {
		favoriteSubmitting = true;
		return async ({
			update
		}: {
			update: (options?: { reset?: boolean; invalidateAll?: boolean }) => Promise<void>;
		}) => {
			await update();
			favoriteSubmitting = false;
		};
	}

	async function handleSendEnhance() {
		sendSubmitting = true;
		return async ({
			update
		}: {
			update: (options?: { reset?: boolean; invalidateAll?: boolean }) => Promise<void>;
		}) => {
			await update();
			sendSubmitting = false;
		};
	}
</script>

<div>
	<form method="POST" action="/articles/{article.id}?/favorite" use:enhance={handleFavoriteEnhance}>
		<button type="submit" aria-busy={favoriteSubmitting}
			>{article.favorite ? 'Unfavorite' : 'Favorite'}</button
		>
	</form>

	{#if canSendToDevice}
		<form method="POST" action="/articles/{article.id}?/send" use:enhance={handleSendEnhance}>
			<button type="submit" aria-busy={sendSubmitting}>Send</button>
		</form>
	{/if}

	<form method="POST" action="/articles/{article.id}?/delete" use:enhance={handleDeleteEnhance}>
		<button type="submit" aria-busy={deleting}>Delete</button>
	</form>
</div>

<style>
	div {
		display: flex;
		gap: 1rem;
		justify-content: center;
	}
</style>
