<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	type NavLink = {
		path: string;
		queryParams?: Record<string, string>;
		label: string;
	};

	const links: NavLink[] = [
		{ path: '/new', label: 'New' },
		{ path: '/articles', label: 'Articles' },
		{ path: '/articles', queryParams: { favorite: 'true' }, label: 'Favorites' },
		{ path: '/account', label: 'Account' }
	] as const;

	const linkHref = (link: NavLink): string => {
		if (!link.queryParams) return link.path;
		const params = new URLSearchParams(link.queryParams);
		return `${link.path}?${params.toString()}` as unknown as '/';
	};

	const isActive = (link: NavLink): boolean => {
		const currentPath = page.url.pathname;
		const currentParams = new URLSearchParams(page.url.search);

		// Path must match exactly
		if (currentPath !== link.path) return false;

		// If link has queryParams, all specified ones must match
		// Extra params in current URL are OK
		if (link.queryParams) {
			for (const [key, value] of Object.entries(link.queryParams)) {
				if (currentParams.get(key) !== value) return false;
			}
			return true;
		}

		// If link has NO queryParams, ensure no conflicting params are set
		// Specifically for /articles links, don't match if favorite=true is set
		if (link.path === '/articles' && currentParams.has('favorite')) {
			return false;
		}

		return true;
	};
</script>

<nav>
	<ul>
		{#each links as link (link.path + JSON.stringify(link.queryParams ?? {}))}
			<li>
				<a
					href={resolve(linkHref(link) as unknown as '/')}
					aria-current={isActive(link) ? 'page' : undefined}
				>
					{link.label}
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
		font-size: 1.3rem;
		text-decoration: none;
	}

	a:not([aria-current='page']) {
		color: var(--pico-secondary);
	}
</style>
