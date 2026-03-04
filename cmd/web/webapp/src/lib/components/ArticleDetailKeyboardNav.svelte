<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';

	let { articleID }: { articleID: string } = $props();
	let favoriteForm: HTMLFormElement;
	let sendForm: HTMLFormElement;
	let deleteForm: HTMLFormElement;

	function handleKeydown(e: KeyboardEvent) {
		switch (e.key) {
			case 'f':
				e.preventDefault();
				toggleFavorite();
				break;
			case 'd':
				e.preventDefault();
				deleteArticle();
				break;
			case 's':
				e.preventDefault();
				sendArticle();
				break;
			case 'ArrowLeft':
			case 'Escape':
				e.preventDefault();
				goto(resolve('/'));
				break;
			case 'n':
				e.preventDefault();
				goto(resolve('/new'));
				break;
		}
	}

	function toggleFavorite() {
		favoriteForm.requestSubmit();
	}

	async function deleteArticle() {
		if (!window.confirm('Are you sure you want to delete this article?')) {
			return;
		}
		deleteForm.requestSubmit();
	}

	function sendArticle() {
		sendForm.requestSubmit();
	}

	onMount(() => {
		document.addEventListener('keydown', handleKeydown);
		return () => document.removeEventListener('keydown', handleKeydown);
	});
</script>

<form
	bind:this={favoriteForm}
	method="POST"
	action="/articles/{articleID}?/favorite"
	use:enhance
></form>
<form bind:this={sendForm} method="POST" action="/articles/{articleID}?/send" use:enhance></form>
<form
	bind:this={deleteForm}
	method="POST"
	action="/articles/{articleID}?/delete"
	use:enhance
></form>
