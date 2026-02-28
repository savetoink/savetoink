<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { enhance } from '$app/forms';
	import ArticleMetaItem from '$lib/components/ArticleMetaItem.svelte';
	import Navigator from '$lib/components/Navigator.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	let selectedArticleIndex = $state<number | null>(null);
	let articleElements: HTMLLIElement[] = [];
	let favoriteForm: HTMLFormElement | undefined;
	let sendForm: HTMLFormElement | undefined;
	let deleteForm: HTMLFormElement | undefined;

	function setArticleElement(element: HTMLLIElement | undefined, index: number) {
		if (element) {
			articleElements[index] = element;
		}
	}

	function toggleFavorite() {
		if (selectedArticleIndex !== null && favoriteForm) {
			favoriteForm.requestSubmit();
		}
	}

	async function deleteArticle() {
		if (selectedArticleIndex === null) return;
		if (!deleteForm) return;
		if (!window.confirm('Are you sure you want to delete this article?')) {
			return;
		}
		deleteForm.requestSubmit();
	}

	function sendArticle() {
		if (selectedArticleIndex !== null && sendForm) {
			sendForm.requestSubmit();
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (!data.articles.length) return;

		switch (e.key) {
			case 'ArrowUp':
			case 'k':
				e.preventDefault();
				if (selectedArticleIndex === null || selectedArticleIndex > 0) {
					selectedArticleIndex = selectedArticleIndex === null ? 0 : selectedArticleIndex - 1;
				}
				break;
			case 'ArrowDown':
			case 'j':
				e.preventDefault();
				if (selectedArticleIndex === null || selectedArticleIndex < data.articles.length - 1) {
					selectedArticleIndex = selectedArticleIndex === null ? 0 : selectedArticleIndex + 1;
				}
				break;
			case 'ArrowRight':
			case 'Enter':
				e.preventDefault();
				if (selectedArticleIndex !== null) {
					window.location.href = `/articles/${data.articles[selectedArticleIndex].id}`;
				}
				break;
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
			case 'n':
				e.preventDefault();
				goto(resolve('/new'));
				break;
		}
	}

	$effect(() => {
		const index = selectedArticleIndex;
		if (index !== null && articleElements[index]) {
			tick().then(() => {
				if (articleElements[index]) {
					articleElements[index].scrollIntoView({ behavior: 'smooth', block: 'start' });
				}
			});
		}
	});

	onMount(() => {
		document.addEventListener('keydown', handleKeydown);
		return () => document.removeEventListener('keydown', handleKeydown);
	});
</script>

{#if data.articles.length === 0}
	<p>No articles found</p>
{:else}
	<ul>
		{#each data.articles as article, index (article.id)}
			<ArticleMetaItem
				{article}
				selected={selectedArticleIndex === index}
				ref={(el: HTMLLIElement) => setArticleElement(el, index)}
			/>
		{/each}
	</ul>
{/if}

<Navigator page={data.page} has_more={data.has_more} />

<form
	bind:this={favoriteForm}
	method="POST"
	action={selectedArticleIndex !== null
		? `/articles/${data.articles[selectedArticleIndex].id}?/favorite`
		: ''}
	use:enhance
></form>
<form
	bind:this={sendForm}
	method="POST"
	action={selectedArticleIndex !== null
		? `/articles/${data.articles[selectedArticleIndex].id}?/send`
		: ''}
	use:enhance
></form>
<form
	bind:this={deleteForm}
	method="POST"
	action={selectedArticleIndex !== null
		? `/articles/${data.articles[selectedArticleIndex].id}?/delete`
		: ''}
	use:enhance
></form>

<style>
	ul {
		padding-left: 0;
	}
</style>
