<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { enhance } from '$app/forms';
	import ArticleControls from '$lib/components/ArticleControls.svelte';
	import KeyboardNav from '$lib/components/KeyboardNav.svelte';
	import ArticleMetaAccordion from '$lib/components/ArticleMetaAccordion.svelte';
	import Tags from '$lib/components/Tags.svelte';
	import TagInput from '$lib/components/TagInput.svelte';

	import {
		DETAIL_BINDINGS,
		toggleFavorite as toggleFavoriteAction,
		deleteArticle as deleteArticleAction,
		sendArticle as sendArticleAction
	} from '@savetoink/shared';
	import type { Article, UserProfile } from '@savetoink/shared';

	type ArticlePageData = Article & { user: UserProfile };

	let { data }: { data: ArticlePageData } = $props();
	const title = $derived(data.title || data.url);

	let favoriteForm = $state<HTMLFormElement>();
	let sendForm = $state<HTMLFormElement>();
	let deleteForm = $state<HTMLFormElement>();
	let addTagsForm = $state<HTMLFormElement>();
	let removeTagsForm = $state<HTMLFormElement>();

	let favoriteSubmitting = $state(false);
	let sendSubmitting = $state(false);
	let deleteSubmitting = $state(false);

	let tagError = $state<string | null>(null);
	let tagsInputValue = $state('');

	const keyboardCallbacks = {
		f: () => favoriteForm && toggleFavoriteAction(favoriteForm),
		d: () => deleteForm && deleteArticleAction(deleteForm),
		s: () => sendForm && sendArticleAction(sendForm),
		ArrowLeft: () => goto(resolve('/articles')),
		Escape: () => goto(resolve('/articles')),
		h: () => goto(resolve('/articles')),
		n: () => goto(resolve('/new')),
		a: () => goto(resolve('/account'))
	};

	const enabledKeys = $derived(
		Object.keys(DETAIL_BINDINGS).filter((key) => key !== 's' || data.user?.device_email)
	);

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

	async function handleDeleteEnhance() {
		deleteSubmitting = true;
		return async ({
			update
		}: {
			update: (options?: { reset?: boolean; invalidateAll?: boolean }) => Promise<void>;
		}) => {
			await update();
			deleteSubmitting = false;
		};
	}

	async function handleAddTagsEnhance() {
		tagError = null;
		return async ({
			update,
			result
		}: {
			update: (options?: { reset?: boolean; invalidateAll?: boolean }) => Promise<void>;
			result: unknown;
		}) => {
			if (result && typeof result === 'object' && 'type' in result) {
				if (result.type === 'error' && 'data' in result) {
					const errorData = result.data as { message?: string } | undefined;
					tagError = errorData?.message || 'Failed to add tags';
				} else if (result.type === 'success') {
					await update();
				}
			}
		};
	}

	async function handleRemoveTagsEnhance() {
		tagError = null;
		return async ({
			update,
			result
		}: {
			update: (options?: { reset?: boolean; invalidateAll?: boolean }) => Promise<void>;
			result: unknown;
		}) => {
			if (result && typeof result === 'object' && 'type' in result) {
				if (result.type === 'error' && 'data' in result) {
					const errorData = result.data as { message?: string } | undefined;
					tagError = errorData?.message || 'Failed to remove tags';
				} else if (result.type === 'success') {
					await update();
				}
			}
		};
	}

	async function handleAddTags(tags: string[]) {
		if (!addTagsForm) return;
		const input = addTagsForm.querySelector('input[name="tags"]') as HTMLInputElement;
		if (input) {
			input.value = tags.join(',');
			addTagsForm.requestSubmit();
		}
	}

	async function handleRemoveTag(tag: string) {
		if (!removeTagsForm) return;
		const input = removeTagsForm.querySelector('input[name="tags"]') as HTMLInputElement;
		if (input) {
			input.value = tag;
			removeTagsForm.requestSubmit();
		}
	}
</script>

<KeyboardNav bindings={DETAIL_BINDINGS} callbacks={keyboardCallbacks} {enabledKeys} />

<article>
	<header>
		<h1>
			<a
				href={data.url}
				target="_blank"
				rel="external noopener"
				title="Open the original link"
				data-tooltip="Open the original link"
				>{#if data.favorite}
					<span>⭐️&nbsp;</span>
				{/if}
				{#if data.author}
					{data.author} -
				{/if}
				{title}</a
			>
			{#if data.imageUrl}
				<picture>
					<img src={data.imageUrl} alt={data.title} />
				</picture>
			{/if}
		</h1>

		<Tags tags={data.tags} />
		<ArticleMetaAccordion article={data} />
		<ArticleControls
			article={data}
			user={data.user}
			{favoriteForm}
			{sendForm}
			{deleteForm}
			{favoriteSubmitting}
			{sendSubmitting}
			{deleteSubmitting}
		/>
	</header>
	<section>
		<!-- eslint-disable-next-line svelte/no-at-html-tags -->
		{@html data.content}
	</section>
</article>

<section class="tag-management">
	<h2>Manage Tags</h2>
	{#if tagError}
		<p class="error" role="alert">{tagError}</p>
	{/if}
	<TagInput tags={data.tags || []} onAdd={handleAddTags} onRemove={handleRemoveTag} />
</section>

<form
	bind:this={favoriteForm}
	method="POST"
	action="/articles/{data.id}?/favorite"
	use:enhance={handleFavoriteEnhance}
></form>
<form
	bind:this={sendForm}
	method="POST"
	action="/articles/{data.id}?/send"
	use:enhance={handleSendEnhance}
></form>
<form
	bind:this={deleteForm}
	method="POST"
	action="/articles/{data.id}?/delete"
	use:enhance={handleDeleteEnhance}
></form>
<form
	bind:this={addTagsForm}
	method="POST"
	action="/articles/{data.id}?/addTags"
	use:enhance={handleAddTagsEnhance}
>
	<input type="hidden" name="tags" bind:value={tagsInputValue} />
</form>
<form
	bind:this={removeTagsForm}
	method="POST"
	action="/articles/{data.id}?/removeTags"
	use:enhance={handleRemoveTagsEnhance}
>
	<input type="hidden" name="tags" bind:value={tagsInputValue} />
</form>

<style>
	img {
		padding-top: 1rem;
	}

	.tag-management {
		margin: 2rem 0;
		padding: 1.5rem;
		border: 1px solid var(--pico-muted-color);
		border-radius: var(--pico-border-radius);
	}

	.tag-management h2 {
		margin-top: 0;
	}

	.tag-management .error {
		color: var(--pico-del-color);
		margin-bottom: 1rem;
	}
</style>
