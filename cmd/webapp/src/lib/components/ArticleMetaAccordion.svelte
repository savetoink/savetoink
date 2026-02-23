<script lang="ts">
	import { formatDate } from '$lib/utils/date';
	let { article } = $props();
</script>

<details>
	<summary class="outline">Show article details</summary>
	<dl>
		{#if article.title}
			<dt>Title</dt>
			<dd>{article.title}</dd>
		{/if}
		{#if article.author}
			<dt>Author</dt>
			<dd>{article.author}</dd>
		{/if}
		{#if article.favorite}
			<dt>Favorite</dt>
			<dd>{article.favorite ? 'Yes' : 'No'}</dd>
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
		<dt>Original link</dt>
		<dd>
			<a href={article.url} target="_blank" rel="external noreferrer">{article.url}</a>
		</dd>
	</dl>
	{#if article.error}
		<p class="error" role="alert">error: {article.error}</p>
	{/if}
</details>
