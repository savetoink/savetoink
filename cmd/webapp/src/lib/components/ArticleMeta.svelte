<script lang="ts">
	import { resolve } from '$app/paths';
	import { formatDate } from '$lib/utils/date';
	import ArticleControls from './ArticleControls.svelte';
	import type { Article } from '$lib/server/types';
	let { article, mode = 'header' }: { article: Article; mode?: 'card' | 'header' } = $props();
</script>

{#if mode === 'card'}
	<article>
		{#if article.title}
			<h2><a href={resolve(`/articles/${article.id}`)}>{article.title}</a></h2>
		{:else}
			<h2><a href={resolve(`/articles/${article.id}`)}>{article.url}</a></h2>
		{/if}
		{#if article.imageUrl}
			<img src={article.imageUrl} alt={article.title} width="30%" />
		{/if}
		{#if article.excerpt}
			<p>{article.excerpt}</p>
		{/if}
		<p>
			Original link: <a href={article.url} target="_blank" rel="external">{article.url}</a>
		</p>
		{#if article.author}
			<p>by {article.author}</p>
		{/if}
		{#if article.readingTimeMinutes}
			<p>{article.readingTimeMinutes} min read</p>
		{/if}
		{#if article.createdAt}
			<p>added: {formatDate(article.createdAt)}</p>
		{/if}
		{#if article.publishedAt}
			<p>published: {formatDate(article.publishedAt)}</p>
		{/if}
		{#if article.deliveryStatus}
			<p>status: {article.deliveryStatus}</p>
		{/if}
		{#if article.error}
			<p class="error">error: {article.error}</p>
		{/if}
		<ArticleControls {article} />
	</article>
{:else if mode === 'header'}
	<header>
		{#if article.title}
			<h1>{article.title}</h1>
		{:else}
			<h1>{article.url}</h1>
		{/if}
		{#if article.author}
			<p>by {article.author}</p>
		{/if}
		{#if article.siteName}
			<p>source: {article.siteName}</p>
		{/if}
		{#if article.sourceDomain}
			<p>domain: {article.sourceDomain}</p>
		{/if}
		{#if article.publishedAt}
			<p>published: {formatDate(article.publishedAt)}</p>
		{/if}
		{#if article.wordCount}
			<p>{article.wordCount} words</p>
		{/if}
		{#if article.readingTimeMinutes}
			<p>{article.readingTimeMinutes} min read</p>
		{/if}
		{#if article.deliveryStatus}
			<p>status: {article.deliveryStatus}</p>
		{/if}
		{#if article.deliveredFrom}
			<p>delivered from: {article.deliveredFrom}</p>
		{/if}
		{#if article.deliveredTo}
			<p>delivered to: {article.deliveredTo}</p>
		{/if}
		{#if article.deliveredBy}
			<p>delivered by: {article.deliveredBy}</p>
		{/if}
		{#if article.deliveredEmailUUID}
			<p>email id: {article.deliveredEmailUUID}</p>
		{/if}
		{#if article.createdAt}
			<p>added: {formatDate(article.createdAt)}</p>
		{/if}
		{#if article.error}
			<p class="error">error: {article.error}</p>
		{/if}
	</header>
{/if}
