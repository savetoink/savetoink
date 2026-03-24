<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	const links = [
		{ href: '/new', label: 'New' },
		{ href: '/articles', label: 'Articles' },
		{ href: '/articles?favorites=true', label: 'Favorites' },
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
	<ul class="tab-strip">
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
		display: flex;
		justify-content: center;
		padding: 1.5rem 0;
	}

	.tab-strip {
		max-height: 2rem;
		display: flex;
		gap: 4px;
		background: rgba(255, 255, 255, 0.06);
		padding: 4px;
		border-radius: 10px;
	}

	a {
		padding: 7px 18px;
		border-radius: 7px;
		font-size: 14px;
		font-weight: 500;
		color: rgba(255, 255, 255, 0.45);
		text-decoration: none;
		transition:
			background 0.15s,
			color 0.15s;
	}

	a:hover:not([aria-current='page']) {
		color: rgba(255, 255, 255, 0.75);
	}

	a[aria-current='page'] {
		background: rgba(255, 255, 255, 0.1);
		color: white;
	}
</style>
