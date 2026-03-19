<script lang="ts">
	import { enhance } from '$app/forms';
	import type { Article, UserProfile } from '@savetoink/shared';

	let {
		article,
		user,
		favoriteSubmitting = false,
		sendSubmitting = false,
		deleteSubmitting = false,
		favoriteEnhance,
		sendEnhance,
		deleteEnhance,
		bind: controls
	}: {
		article: Article;
		user?: UserProfile;
		favoriteSubmitting?: boolean;
		sendSubmitting?: boolean;
		deleteSubmitting?: boolean;
		favoriteEnhance?: ReturnType<typeof enhance>;
		sendEnhance?: ReturnType<typeof enhance>;
		deleteEnhance?: ReturnType<typeof enhance>;
		bind?: {
			favoriteForm: HTMLFormElement;
			sendForm: HTMLFormElement;
			deleteForm: HTMLFormElement;
		};
	} = $props();

	let favoriteForm: HTMLFormElement;
	let sendForm: HTMLFormElement;
	let deleteForm: HTMLFormElement;

	$effect(() => {
		if (controls && favoriteForm) {
			controls.favoriteForm = favoriteForm;
		}
		if (controls && sendForm) {
			controls.sendForm = sendForm;
		}
		if (controls && deleteForm) {
			controls.deleteForm = deleteForm;
		}
	});

	const canSendToDevice = $derived(!!user?.device_email);

	async function handleDeleteEnhance({ cancel }: { cancel: () => void }) {
		if (!window.confirm('Are you sure you want to delete this article?')) {
			cancel();
		}
	}

	const deleteEnhanceFn = $derived(deleteEnhance || handleDeleteEnhance);
</script>

<div>
	<form bind:this={favoriteForm} method="POST" action="/articles/{article.id}?/favorite" use:enhance={favoriteEnhance}>
		<button type="submit" aria-busy={favoriteSubmitting}
			>{article.favorite ? 'Unfavorite' : 'Favorite'}</button
		>
	</form>

	{#if canSendToDevice}
		<form bind:this={sendForm} method="POST" action="/articles/{article.id}?/send" use:enhance={sendEnhance}>
			<button type="submit" aria-busy={sendSubmitting}>Send</button>
		</form>
	{/if}

	<form bind:this={deleteForm} method="POST" action="/articles/{article.id}?/delete" use:enhance={deleteEnhanceFn}>
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
