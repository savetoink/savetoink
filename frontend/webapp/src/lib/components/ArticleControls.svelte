<script lang="ts">
	import { enhance } from '$app/forms';
	import type { Article, UserProfile } from '@savetoink/shared';

	let { article, user }: { article: Article; user?: UserProfile } = $props();

	const canSendToDevice = $derived(!!user?.device_email);

	async function handleEnhance({ cancel }: { cancel: () => void }) {
		if (!window.confirm('Are you sure you want to delete this article?')) {
			cancel();
		}
	}
</script>

<div>
	<form method="POST" action="/articles/{article.id}?/favorite" use:enhance>
		<button type="submit">{article.favorite ? 'Unfavorite' : 'Favorite'}</button>
	</form>

	{#if canSendToDevice}
		<form method="POST" action="/articles/{article.id}?/send">
			<button type="submit">Send</button>
		</form>
	{/if}

	<form method="POST" action="/articles/{article.id}?/delete" use:enhance={handleEnhance}>
		<button type="submit">Delete</button>
	</form>
</div>

<style>
	div {
		display: flex;
		gap: 1rem;
		justify-content: center;
	}
</style>
