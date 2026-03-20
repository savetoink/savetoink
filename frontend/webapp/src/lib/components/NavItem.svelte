<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	let { link, text } = $props();

	const linkUrl = $derived(
		new URL(link, typeof window !== 'undefined' ? window.location.origin : 'http://example.com')
	);
</script>

<li>
	<a
		href={resolve(link)}
		aria-current={page.url.pathname === linkUrl.pathname && page.url.search === linkUrl.search
			? 'page'
			: undefined}
	><span>{text}</span></a
	>
</li>

<style>
	li {
		flex: 1;
	}

	a {
		display: flex;
		justify-content: center;
		align-items: center;
		padding: 0.75rem 1.5rem;
		background-color: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		color: #374151;
		text-decoration: none;
		transition: all 0.2s;
		font-weight: 500;
		height: 100%;
	}

	a:hover {
		background-color: #f9fafb;
		border-color: #d1d5db;
	}

	a[aria-current='page'] {
		background-color: #3b82f6;
		color: #fff;
		border-color: #3b82f6;
	}
</style>
