<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	let { link, text } = $props();
</script>

<li>
	<a href={resolve(link)} aria-current={page.url.pathname === link ? 'page' : undefined}>{text}</a>
</li>
