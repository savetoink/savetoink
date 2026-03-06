<script lang="ts">
	import ArticleControls from '$lib/components/ArticleControls.svelte';
	import ArticleDetailKeyboardNav from '$lib/components/ArticleDetailKeyboardNav.svelte';
	import ArticleMetaAccordion from '$lib/components/ArticleMetaAccordion.svelte';
	import type { Article, UserProfile } from '@savetoink/shared';

	type ArticlePageData = Article & { user: UserProfile };

	let { data }: { data: ArticlePageData } = $props();
	const title = $derived(data.title || data.url);
</script>

<ArticleDetailKeyboardNav articleID={data.id} />

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
				{title}</a
			>
		</h1>

		<ArticleMetaAccordion article={data} />
		<ArticleControls article={data} user={data.user} />
	</header>
	<section>
		<!-- eslint-disable-next-line svelte/no-at-html-tags -->
		{@html data.content}
	</section>
</article>
