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

	const keyboardCallbacks = {
		f: () => toggleFavoriteAction(favoriteForm),
		d: () => deleteArticleAction(deleteForm),
		s: () => sendArticleAction(sendForm),
		ArrowLeft: () => goto(resolve('/articles')),
		Escape: () => goto(resolve('/articles')),
		h: () => goto(resolve('/articles')),
		n: () => goto(resolve('/new')),
		a: () => goto(resolve('/account'))
	};

	const enabledKeys = $derived(
		Object.keys(DETAIL_BINDINGS).filter((key) => key !== 's' || data.user?.device_email)
	);
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
		<ArticleControls article={data} user={data.user} />
	</header>
	<section>
		<!-- eslint-disable-next-line svelte/no-at-html-tags -->
		{@html data.content}
	</section>
</article>

<form
	bind:this={favoriteForm}
	method="POST"
	action="/articles/{data.id}?/favorite"
	use:enhance
></form>
<form bind:this={sendForm} method="POST" action="/articles/{data.id}?/send" use:enhance></form>
<form bind:this={deleteForm} method="POST" action="/articles/{data.id}?/delete" use:enhance></form>

<style>
	img {
		padding-top: 1rem;
	}
</style>
