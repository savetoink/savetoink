<script lang="ts">
	import { tick } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { enhance } from '$app/forms';
	import ArticleMetaItem from '$lib/components/ArticleMetaItem.svelte';
	import Navigator from '$lib/components/Navigator.svelte';
	import KeyboardNav from '$lib/components/KeyboardNav.svelte';
	import {
		LIST_BINDINGS,
		toggleFavorite as toggleFavoriteAction,
		deleteArticle as deleteArticleAction,
		sendArticle as sendArticleAction
	} from '@savetoink/shared';
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
			toggleFavoriteAction(favoriteForm);
		}
	}

	async function deleteArticle() {
		if (selectedArticleIndex === null) return;
		if (!deleteForm) return;
		deleteArticleAction(deleteForm);
	}

	function sendArticle() {
		if (selectedArticleIndex !== null && sendForm) {
			sendArticleAction(sendForm);
		}
	}

	const keyboardCallbacks = {
		ArrowUp: () => {
			if (selectedArticleIndex === null || selectedArticleIndex > 0) {
				selectedArticleIndex = selectedArticleIndex === null ? 0 : selectedArticleIndex - 1;
			}
		},
		k: () => {
			if (selectedArticleIndex === null || selectedArticleIndex > 0) {
				selectedArticleIndex = selectedArticleIndex === null ? 0 : selectedArticleIndex - 1;
			}
		},
		ArrowDown: () => {
			if (selectedArticleIndex === null || selectedArticleIndex < data.articles.length - 1) {
				selectedArticleIndex = selectedArticleIndex === null ? 0 : selectedArticleIndex + 1;
			}
		},
		j: () => {
			if (selectedArticleIndex === null || selectedArticleIndex < data.articles.length - 1) {
				selectedArticleIndex = selectedArticleIndex === null ? 0 : selectedArticleIndex + 1;
			}
		},
		ArrowRight: () => {
			if (selectedArticleIndex !== null) {
				window.location.href = `/articles/${data.articles[selectedArticleIndex].id}`;
			}
		},
		Enter: () => {
			if (selectedArticleIndex !== null) {
				window.location.href = `/articles/${data.articles[selectedArticleIndex].id}`;
			}
		},
		f: () => toggleFavorite(),
		d: () => deleteArticle(),
		s: () => sendArticle(),
		n: () => goto(resolve('/new')),
		h: () => goto(resolve('/')),
		a: () => goto(resolve('/account'))
	};

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
</script>

{#if data.articles.length === 0}
	<p>No articles found</p>
{:else}
	<ul>
		{#each data.articles as article, index (article.id)}
			<ArticleMetaItem
				{article}
				user={data.user}
				selected={selectedArticleIndex === index}
				ref={(el: HTMLLIElement) => setArticleElement(el, index)}
			/>
		{/each}
	</ul>
{/if}

<Navigator page={data.page} hasMore={data.has_more} />

<KeyboardNav bindings={LIST_BINDINGS} callbacks={keyboardCallbacks} />

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
