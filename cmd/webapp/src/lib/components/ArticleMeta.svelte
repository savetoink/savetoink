<script lang="ts">
	import { resolve } from '$app/paths';
	import { formatDate } from '$lib/utils/date';
	import ArticleControls from './ArticleControls.svelte';
	let { article, mode = 'header' } = $props();
</script>

{#if mode === 'card'}
	<article>
		<header>
			{#if article.title}
				<h2><a href={resolve(`/articles/${article.id}`)}>{article.title}</a></h2>
			{:else}
				<h2><a href={resolve(`/articles/${article.id}`)}>{article.url}</a></h2>
			{/if}
			{#if article.excerpt}
				<p>{article.excerpt}</p>
			{/if}
		</header>
		{#if article.imageUrl}
			<figure>
				<img src={article.imageUrl} alt={article.title || 'Article image'} />
			</figure>
		{/if}
		<footer>
			<dl>
				{#if article.author}
					<dt>Author</dt>
					<dd>{article.author}</dd>
				{/if}
				{#if article.readingTimeMinutes}
					<dt>Reading time</dt>
					<dd>{article.readingTimeMinutes} min</dd>
				{/if}
				{#if article.createdAt}
					<dt>Added</dt>
					<dd><time datetime={article.createdAt}>{formatDate(article.createdAt)}</time></dd>
				{/if}
				{#if article.publishedAt}
					<dt>Published</dt>
					<dd><time datetime={article.publishedAt}>{formatDate(article.publishedAt)}</time></dd>
				{/if}
				{#if article.deliveryStatus}
					<dt>Status</dt>
					<dd>{article.deliveryStatus}</dd>
				{/if}
				{#if article.tags}
					<dt>Tags</dt>
					<dd>{article.tags.join(', ')}</dd>
				{/if}
			</dl>
			{#if article.error}
				<p class="error" role="alert">error: {article.error}</p>
			{/if}
			<p>
				<a href={article.url} target="_blank" rel="external noreferrer">Original link</a>
			</p>
			<ArticleControls {article} />
		</footer>
	</article>
{:else if mode === 'header'}
	<header>
		{#if article.title}
			<h1>{article.title}</h1>
		{:else}
			<h1>{article.url}</h1>
		{/if}
		<dl>
			{#if article.author}
				<dt>Author</dt>
				<dd>{article.author}</dd>
			{/if}
			{#if article.siteName}
				<dt>Source</dt>
				<dd>{article.siteName}</dd>
			{/if}
			{#if article.sourceDomain}
				<dt>Domain</dt>
				<dd>{article.sourceDomain}</dd>
			{/if}
			{#if article.publishedAt}
				<dt>Published</dt>
				<dd><time datetime={article.publishedAt}>{formatDate(article.publishedAt)}</time></dd>
			{/if}
			{#if article.wordCount}
				<dt>Word count</dt>
				<dd>{article.wordCount} words</dd>
			{/if}
			{#if article.readingTimeMinutes}
				<dt>Reading time</dt>
				<dd>{article.readingTimeMinutes} min</dd>
			{/if}
			{#if article.deliveryStatus}
				<dt>Status</dt>
				<dd>{article.deliveryStatus}</dd>
			{/if}
			{#if article.deliveredFrom}
				<dt>Delivered from</dt>
				<dd>{article.deliveredFrom}</dd>
			{/if}
			{#if article.deliveredTo}
				<dt>Delivered to</dt>
				<dd>{article.deliveredTo}</dd>
			{/if}
			{#if article.deliveredBy}
				<dt>Delivered by</dt>
				<dd>{article.deliveredBy}</dd>
			{/if}
			{#if article.deliveredEmailUUID}
				<dt>Email ID</dt>
				<dd>{article.deliveredEmailUUID}</dd>
			{/if}
			{#if article.createdAt}
				<dt>Added</dt>
				<dd><time datetime={article.createdAt}>{formatDate(article.createdAt)}</time></dd>
			{/if}
			{#if article.tags}
				<dt>Tags</dt>
				<dd>{article.tags.join(', ')}</dd>
			{/if}
		</dl>
		<p>
			Original link: <a href={article.url} target="_blank" rel="external noreferrer"
				>{article.url}</a
			>
		</p>
		{#if article.error}
			<p class="error" role="alert">error: {article.error}</p>
		{/if}
	</header>
{/if}
