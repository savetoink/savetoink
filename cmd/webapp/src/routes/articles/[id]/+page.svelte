<script lang="ts">
	import type { PageData } from './$types';
	import ArticleControls from '$lib/components/ArticleControls.svelte';
	import ArticleMeta from '$lib/components/ArticleMeta.svelte';

	let { data }: { data: PageData } = $props();
</script>

<ArticleControls article={data} />
<ArticleMeta article={data} mode="header" />
<!-- eslint-disable-next-line svelte/no-at-html-tags -->
{@html data.content}
