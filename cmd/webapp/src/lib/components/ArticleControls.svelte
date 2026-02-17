<script lang="ts">
	import { enhance } from '$app/forms';
	let { article } = $props();

	async function handleEnhance({ cancel }: { cancel: () => void }) {
		if (!window.confirm('Are you sure you want to delete this article?')) {
			cancel();
		}
	}
</script>

<form method="POST" action="/articles/{article.id}?/delete" use:enhance={handleEnhance}>
	<button type="submit">Delete</button>
</form>
