<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	const links = [
		{ href: '/new', label: 'New' },
		{ href: '/articles', label: 'Articles' },
		{ href: '/articles?favorite=true', label: 'Favorites' },
		{ href: '/account', label: 'Account' }
	] as const;

	const isActive = (href: string) => {
		const currentPath = page.url.pathname;
		const currentParams = new URLSearchParams(page.url.search);
		const [linkPath, linkParams] = href.split('?');
		const params = new URLSearchParams(linkParams);

		if (currentPath !== linkPath) return false;

		for (const [key, value] of params) {
			if (currentParams.get(key) !== value) return false;
		}

		for (const [key, value] of currentParams) {
			if (params.get(key) !== value) return false;
		}

		return true;
	};
</script>

<nav>
	<ul>
		{#each links as { href, label } (href)}
			<li>
				<a href={resolve(href)} aria-current={isActive(href) ? 'page' : undefined}>
					{label}
				</a>
			</li>
		{/each}
	</ul>
</nav>

<style>
	nav {
		justify-content: center;
		padding: 1rem 0;
	}

	a {
		background: var(--pico-switch-background-color);
		text-decoration: none;
	}

	a:not([aria-current='page']) {
		color: white;
	}
</style>
