<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { enhance } from '$app/forms';
	import ArticleControls from '$lib/components/ArticleControls.svelte';
	import KeyboardNav from '$lib/components/KeyboardNav.svelte';
	import ArticleMetaAccordion from '$lib/components/ArticleMetaAccordion.svelte';
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

	let favoriteForm: HTMLFormElement;
	let sendForm: HTMLFormElement;
	let deleteForm: HTMLFormElement;

	let favoriteSubmitting = $state(false);
	let sendSubmitting = $state(false);
	let deleteSubmitting = $state(false);

	let controls: {
		favoriteForm: HTMLFormElement;
		sendForm: HTMLFormElement;
		deleteForm: HTMLFormElement;
	};

	const keyboardCallbacks = $derived({
		f: () => toggleFavoriteAction(controls?.favoriteForm),
		d: () => deleteArticleAction(controls?.deleteForm),
		s: () => sendArticleAction(controls?.sendForm),
		ArrowLeft: () => goto(resolve('/articles')),
		Escape: () => goto(resolve('/articles')),
		h: () => goto(resolve('/articles')),
		n: () => goto(resolve('/new')),
		a: () => goto(resolve('/account'))
	});

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
		if (!window.confirm('Are you sure you want to delete this article?')) {
			return async ({ cancel }: { cancel: () => void }) => {
				cancel();
			};
		}
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

		<ArticleMetaAccordion article={data} />
		<ArticleControls
			bind:controls
			article={data}
			user={data.user}
			{favoriteSubmitting}
			{sendSubmitting}
			{deleteSubmitting}
			favoriteEnhance={handleFavoriteEnhance()}
			sendEnhance={handleSendEnhance()}
			deleteEnhance={handleDeleteEnhance()}
		/>
	</header>
	<section>
		<!-- eslint-disable-next-line svelte/no-at-html-tags -->
		{@html data.content}
	</section>
</article>

<style>
	img {
		padding-top: 1rem;
	}
</style>
