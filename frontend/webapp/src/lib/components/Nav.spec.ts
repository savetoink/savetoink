import { describe, it, expect } from 'vitest';

// Test the isActive logic directly without browser rendering
describe('Nav.svelte - isActive logic', () => {
	type NavLink = {
		path: string;
		queryParams?: Record<string, string>;
		label: string;
	};

	const isActive = (link: NavLink, currentPath: string, currentSearch: string): boolean => {
		const currentParams = new URLSearchParams(currentSearch);

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

	const links: NavLink[] = [
		{ path: '/new', label: 'New' },
		{ path: '/articles', label: 'Articles' },
		{ path: '/articles', queryParams: { favorite: 'true' }, label: 'Favorites' },
		{ path: '/account', label: 'Account' }
	];

	describe('Articles link (no queryParams)', () => {
		it('should be active on /articles', () => {
			const link = links.find((l) => l.label === 'Articles')!;
			expect(isActive(link, '/articles', '')).toBe(true);
		});

		it('should be active on /articles?page=1', () => {
			const link = links.find((l) => l.label === 'Articles')!;
			expect(isActive(link, '/articles', '?page=1')).toBe(true);
		});

		it('should be active on /articles?tag=tech&page=2', () => {
			const link = links.find((l) => l.label === 'Articles')!;
			expect(isActive(link, '/articles', '?tag=tech&page=2')).toBe(true);
		});

		it('should NOT be active on /articles?favorite=true', () => {
			const link = links.find((l) => l.label === 'Articles')!;
			expect(isActive(link, '/articles', '?favorite=true')).toBe(false);
		});

		it('should NOT be active on /articles?favorite=true&page=2', () => {
			const link = links.find((l) => l.label === 'Articles')!;
			expect(isActive(link, '/articles', '?favorite=true&page=2')).toBe(false);
		});

		it('should NOT be active on /articles?favorite=false', () => {
			const link = links.find((l) => l.label === 'Articles')!;
			expect(isActive(link, '/articles', '?favorite=false')).toBe(false);
		});

		it('should NOT be active on /new', () => {
			const link = links.find((l) => l.label === 'Articles')!;
			expect(isActive(link, '/new', '')).toBe(false);
		});
	});

	describe('Favorites link (with favorite=true)', () => {
		it('should be active on /articles?favorite=true', () => {
			const link = links.find((l) => l.label === 'Favorites')!;
			expect(isActive(link, '/articles', '?favorite=true')).toBe(true);
		});

		it('should be active on /articles?favorite=true&page=1', () => {
			const link = links.find((l) => l.label === 'Favorites')!;
			expect(isActive(link, '/articles', '?favorite=true&page=1')).toBe(true);
		});

		it('should be active on /articles?tag=tech&favorite=true&page=2', () => {
			const link = links.find((l) => l.label === 'Favorites')!;
			expect(isActive(link, '/articles', '?tag=tech&favorite=true&page=2')).toBe(true);
		});

		it('should NOT be active on /articles (no favorite param)', () => {
			const link = links.find((l) => l.label === 'Favorites')!;
			expect(isActive(link, '/articles', '')).toBe(false);
		});

		it('should NOT be active on /articles?favorite=false', () => {
			const link = links.find((l) => l.label === 'Favorites')!;
			expect(isActive(link, '/articles', '?favorite=false')).toBe(false);
		});

		it('should NOT be active on /articles?page=1 (no favorite param)', () => {
			const link = links.find((l) => l.label === 'Favorites')!;
			expect(isActive(link, '/articles', '?page=1')).toBe(false);
		});
	});

	describe('New link', () => {
		it('should be active on /new', () => {
			const link = links.find((l) => l.label === 'New')!;
			expect(isActive(link, '/new', '')).toBe(true);
		});

		it('should be active on /new?ref=home', () => {
			const link = links.find((l) => l.label === 'New')!;
			expect(isActive(link, '/new', '?ref=home')).toBe(true);
		});

		it('should NOT be active on /articles', () => {
			const link = links.find((l) => l.label === 'New')!;
			expect(isActive(link, '/articles', '')).toBe(false);
		});
	});

	describe('Account link', () => {
		it('should be active on /account', () => {
			const link = links.find((l) => l.label === 'Account')!;
			expect(isActive(link, '/account', '')).toBe(true);
		});

		it('should NOT be active on /articles', () => {
			const link = links.find((l) => l.label === 'Account')!;
			expect(isActive(link, '/articles', '')).toBe(false);
		});
	});

	describe('Mutual exclusivity', () => {
		it('should have only Favorites active, not Articles, when on /articles?favorite=true', () => {
			const articlesLink = links.find((l) => l.label === 'Articles')!;
			const favoritesLink = links.find((l) => l.label === 'Favorites')!;

			expect(isActive(articlesLink, '/articles', '?favorite=true')).toBe(false);
			expect(isActive(favoritesLink, '/articles', '?favorite=true')).toBe(true);
		});

		it('should have only Articles active, not Favorites, when on /articles', () => {
			const articlesLink = links.find((l) => l.label === 'Articles')!;
			const favoritesLink = links.find((l) => l.label === 'Favorites')!;

			expect(isActive(articlesLink, '/articles', '')).toBe(true);
			expect(isActive(favoritesLink, '/articles', '')).toBe(false);
		});

		it('should have only Articles active, not Favorites, when on /articles?page=1', () => {
			const articlesLink = links.find((l) => l.label === 'Articles')!;
			const favoritesLink = links.find((l) => l.label === 'Favorites')!;

			expect(isActive(articlesLink, '/articles', '?page=1')).toBe(true);
			expect(isActive(favoritesLink, '/articles', '?page=1')).toBe(false);
		});

		it('should have only Favorites active, not Articles, when on /articles?tag=tech&favorite=true&page=2', () => {
			const articlesLink = links.find((l) => l.label === 'Articles')!;
			const favoritesLink = links.find((l) => l.label === 'Favorites')!;

			expect(isActive(articlesLink, '/articles', '?tag=tech&favorite=true&page=2')).toBe(false);
			expect(isActive(favoritesLink, '/articles', '?tag=tech&favorite=true&page=2')).toBe(true);
		});
	});
});
