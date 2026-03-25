<script lang="ts">
	import { tick } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { enhance } from '$app/forms';
	import ArticleMetaItem from '$lib/components/ArticleMetaItem.svelte';
	import Paginator from '$lib/components/Paginator.svelte';
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

	let showTagFilter = $state(false);
	let tagSearchInput = $state('');
	let allTags = $state<string[]>([]);

	let currentTag = $derived(
		new URLSearchParams(typeof window !== 'undefined' ? window.location.search : '').get('tag') ||
			null
	);
	let filteredTags = $derived(
		allTags.filter((tag) => tag.toLowerCase().includes(tagSearchInput.toLowerCase()))
	);

	async function loadAllTags() {
		try {
			const response = await fetch('/api/tags');
			if (response.ok) {
				const result = await response.json();
				allTags = result.tags || [];
			}
		} catch (e) {
			console.error('Failed to load tags:', e);
		}
	}

	function filterByTag(tag: string) {
		const url = new URL(window.location.href);
		url.searchParams.set('tag', tag);
		url.searchParams.set('page', '1');
		window.location.href = url.toString();
	}

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
		h: () => goto(resolve('/articles')),
		a: () => goto(resolve('/account'))
	};

	const enabledKeys = $derived(
		Object.keys(LIST_BINDINGS).filter((key) => key !== 's' || data.user?.device_email)
	);

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

	$effect(() => {
		if (showTagFilter && allTags.length === 0) {
			loadAllTags();
		}
	});
</script>

{#if data.articles.length === 0}
	<p>No articles found</p>
{:else}
	<div class="tag-filter">
		<button
			type="button"
			class="outline"
			onclick={() => (showTagFilter = !showTagFilter)}
			aria-expanded={showTagFilter}
		>
			{#if currentTag}
				Filtered by: <strong>#{currentTag}</strong>
			{:else}
				Filter by tag
			{/if}
		</button>

		{#if currentTag}
			<a href={resolve('/articles')} class="clear-filter">Clear filter</a>
		{/if}

		{#if showTagFilter}
			<div class="tag-dropdown">
				<input
					type="text"
					placeholder="Search tags..."
					bind:value={tagSearchInput}
					autocomplete="off"
				/>

				{#if filteredTags.length > 0}
					<ul class="tag-suggestions">
						{#each filteredTags as tag (tag)}
							<li>
								<button
									type="button"
									onclick={() => filterByTag(tag)}
									class={currentTag === tag ? 'active' : ''}
								>
									#{tag}
								</button>
							</li>
						{/each}
					</ul>
				{:else if tagSearchInput}
					<p class="no-results">No tags found</p>
				{:else}
					<p class="no-results">No tags available. Add tags to articles to see them here.</p>
				{/if}
			</div>
		{/if}
	</div>

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

<Paginator page={data.page} hasMore={data.has_more} />

<KeyboardNav bindings={LIST_BINDINGS} callbacks={keyboardCallbacks} {enabledKeys} />

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

	.tag-filter {
		position: relative;
		margin-bottom: 1rem;
		display: flex;
		align-items: center;
		gap: 1rem;
		flex-wrap: wrap;
	}

	.tag-filter button.outline {
		white-space: nowrap;
	}

	.clear-filter {
		font-size: 0.875rem;
		color: var(--muted-color);
		text-decoration: none;
	}

	.clear-filter:hover {
		text-decoration: underline;
		color: var(--pico-primary);
	}

	.tag-dropdown {
		position: absolute;
		top: 100%;
		left: 0;
		z-index: 100;
		background-color: var(--pico-background-color);
		border: 1px solid var(--pico-muted-border-color);
		border-radius: var(--pico-border-radius);
		box-shadow: var(--pico-box-shadow);
		min-width: 300px;
		max-width: 100%;
		padding: 1rem;
		margin-top: 0.5rem;
	}

	.tag-dropdown input {
		width: 100%;
		margin-bottom: 0.5rem;
	}

	.tag-suggestions {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		padding: 0;
		margin: 0;
		max-height: 200px;
		overflow-y: auto;
		list-style: none;
	}

	.tag-suggestions li {
		list-style: none;
	}

	.tag-suggestions button {
		width: 100%;
		text-align: left;
		padding: 0.5rem;
		border: none;
		background: transparent;
		border-radius: var(--pico-border-radius);
		cursor: pointer;
		transition: background-color 0.2s;
	}

	.tag-suggestions button:hover,
	.tag-suggestions button.active {
		background-color: var(--pico-primary-background);
		color: var(--pico-color);
	}

	.no-results {
		padding: 0.5rem;
		margin: 0;
		color: var(--muted-color);
		font-style: italic;
	}
</style>
