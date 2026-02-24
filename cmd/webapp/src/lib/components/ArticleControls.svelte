<script lang="ts">
	import { enhance } from '$app/forms';
	let { article } = $props();

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
