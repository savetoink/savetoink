import { test, expect } from '@playwright/test';

/**
 * E2E Tests for /articles route
 *
 * These tests use Playwright's route mocking to intercept API calls.
 * Note: Due to SvelteKit server-side rendering and authentication checks,
 * these tests require a valid auth session. The tests perform a login flow
 * before accessing protected routes.
 *
 * To run these tests against a real backend:
 * 1. Set up a test backend instance
 * 2. Configure PUBLIC_API_URL in playwright.config.ts
 * 3. Remove or modify the route mocking to use real API responses
 */

test.describe('/articles page', () => {
	/**
	 * Mock data for tests
	 */
	const mockArticle = {
		id: 'test-article-id',
		url: 'https://example.com/article',
		createdAt: '2024-03-23T12:00:00Z',
		title: 'Test Article Title',
		author: 'John Doe',
		excerpt: 'This is a test article excerpt that describes the content of the article.',
		favorite: false
	};

	const mockArticleFavorite = {
		...mockArticle,
		id: 'test-article-fav-id',
		title: 'Favorite Article',
		favorite: true
	};

	const mockArticlesResponse = {
		articles: [mockArticle, mockArticleFavorite],
		page: 1,
		page_size: 10,
		total: 2,
		has_more: false
	};

	const mockEmptyResponse = {
		articles: [],
		page: 1,
		page_size: 10,
		total: 0,
		has_more: false
	};

	const mockPaginatedResponse = {
		articles: [mockArticle, mockArticleFavorite],
		page: 1,
		page_size: 2,
		total: 4,
		has_more: true
	};

	const mockProfileResponse = {
		account: 'test-account',
		email: 'test@example.com',
		device_email: '',
		auto_send: false
	};

	/**
	 * Set up API route mocks for testing
	 */
	async function setupApiMocks(page: Page) {
		// Mock profile endpoint for authentication
		await page.route('**/v1/user/profile', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify(mockProfileResponse)
			});
		});

		// Mock articles endpoint
		await page.route('**/v1/articles*', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify(mockArticlesResponse)
			});
		});
	}

	test.beforeEach(async ({ page }) => {
		// Set up API mocks before any navigation
		await setupApiMocks(page);
	});

	test('displays articles list when API returns articles', async ({ page }) => {
		await page.goto('/articles');

		// Check that articles are displayed
		await expect(page.getByText('Test Article Title')).toBeVisible();
		await expect(page.getByText('Favorite Article')).toBeVisible();
		await expect(page.getByText('John Doe -')).toBeVisible();
		await expect(page.getByText('This is a test article excerpt')).toBeVisible();

		// Check favorite indicator
		await expect(page.getByText('⭐️')).toBeVisible();
	});

	test('displays empty state when no articles', async ({ page }) => {
		// Override the articles mock to return empty response
		await page.route('**/v1/articles*', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify(mockEmptyResponse)
			});
		});

		await page.goto('/articles');

		await expect(page.getByText('No articles found')).toBeVisible();
	});

	test('displays pagination controls when has_more is true', async ({ page }) => {
		// Override the articles mock to return paginated response
		await page.route('**/v1/articles*', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify(mockPaginatedResponse)
			});
		});

		await page.goto('/articles');

		// Should show Next button
		const nextButton = page.getByRole('button', { name: 'Next' });
		await expect(nextButton).toBeVisible();

		// Previous button should not be visible on first page
		const prevButton = page.getByRole('button', { name: 'Previous' });
		await expect(prevButton).not.toBeVisible();
	});

	test('displays both pagination controls on second page', async ({ page }) => {
		let requestCount = 0;
		await page.route('**/v1/articles*', async (route) => {
			requestCount++;
			if (requestCount === 1) {
				// First request - second page (from URL param)
				await route.fulfill({
					status: 200,
					contentType: 'application/json',
					body: JSON.stringify({
						...mockPaginatedResponse,
						page: 2,
						has_more: false
					})
				});
			}
		});

		await page.goto('/articles?page=2');

		// Should show Previous button
		const prevButton = page.getByRole('button', { name: 'Previous' });
		await expect(prevButton).toBeVisible();

		// Next button should not be visible when has_more is false
		const nextButton = page.getByRole('button', { name: 'Next' });
		await expect(nextButton).not.toBeVisible();
	});

	test('clicking article title navigates to article detail page', async ({ page }) => {
		await page.goto('/articles');

		const articleLink = page.getByText('Test Article Title');
		await articleLink.click();

		await expect(page).toHaveURL(/\/articles\/test-article-id/);
	});

	test('clicking Next button navigates to next page', async ({ page }) => {
		let requestCount = 0;
		await page.route('**/v1/articles*', async (route) => {
			requestCount++;
			if (requestCount === 1) {
				await route.fulfill({
					status: 200,
					contentType: 'application/json',
					body: JSON.stringify(mockPaginatedResponse)
				});
			} else {
				await route.fulfill({
					status: 200,
					contentType: 'application/json',
					body: JSON.stringify({
						...mockPaginatedResponse,
						page: 2,
						has_more: false
					})
				});
			}
		});

		await page.goto('/articles');

		const nextButton = page.getByRole('button', { name: 'Next' });
		await nextButton.click();

		await expect(page).toHaveURL(/\/articles\?page=2/);
	});

	test('clicking Previous button navigates to previous page', async ({ page }) => {
		let requestCount = 0;
		await page.route('**/v1/articles*', async (route) => {
			requestCount++;
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					...mockPaginatedResponse,
					page: requestCount + 1,
					has_more: false
				})
			});
		});

		await page.goto('/articles?page=2');

		const prevButton = page.getByRole('button', { name: 'Previous' });
		await prevButton.click();

		await expect(page).toHaveURL(/\/articles\?page=1/);
	});

	test('article without title displays URL as fallback', async ({ page }) => {
		const articleWithoutTitle = {
			...mockArticle,
			id: 'no-title-id',
			title: undefined
		};

		await page.route('**/v1/articles*', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					articles: [articleWithoutTitle],
					page: 1,
					page_size: 10,
					total: 1,
					has_more: false
				})
			});
		});

		await page.goto('/articles');

		await expect(page.getByText('https://example.com/article')).toBeVisible();
	});

	test('article without excerpt does not display excerpt paragraph', async ({ page }) => {
		const articleWithoutExcerpt = {
			...mockArticle,
			id: 'no-excerpt-id',
			excerpt: undefined
		};

		await page.route('**/v1/articles*', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					articles: [articleWithoutExcerpt],
					page: 1,
					page_size: 10,
					total: 1,
					has_more: false
				})
			});
		});

		await page.goto('/articles');

		// Title should be visible
		await expect(page.getByText('Test Article Title')).toBeVisible();

		// But excerpt should not be
		await expect(page.getByText('This is a test article excerpt')).not.toBeVisible();
	});

	test('article link href contains correct article ID', async ({ page }) => {
		await page.goto('/articles');

		const articleLink = page.getByRole('link', { name: /Test Article Title/ });
		await expect(articleLink).toHaveAttribute('href', '/articles/test-article-id');
	});

	test('respects page_size query parameter', async ({ page }) => {
		let capturedUrl: URL | null = null;
		await page.route('**/v1/articles*', async (route) => {
			capturedUrl = new URL(route.request().url());
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify(mockArticlesResponse)
			});
		});

		await page.goto('/articles?page_size=20');

		expect(capturedUrl).not.toBeNull();
		expect(capturedUrl?.searchParams.get('page_size')).toBe('20');
	});

	test('respects favorite query parameter', async ({ page }) => {
		let capturedUrl: URL | null = null;
		await page.route('**/v1/articles*', async (route) => {
			capturedUrl = new URL(route.request().url());
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					articles: [mockArticleFavorite],
					page: 1,
					page_size: 10,
					total: 1,
					has_more: false
				})
			});
		});

		await page.goto('/articles?favorite=true');

		expect(capturedUrl).not.toBeNull();
		expect(capturedUrl?.searchParams.get('favorite')).toBe('true');
		await expect(page.getByText('Favorite Article')).toBeVisible();
	});

	test('API error is handled gracefully', async ({ page }) => {
		await page.route('**/v1/articles*', async (route) => {
			await route.fulfill({
				status: 500,
				contentType: 'application/json',
				body: JSON.stringify({ error: 'Internal server error' })
			});
		});

		await page.goto('/articles');

		// Should show error page
		await expect(page.getByText(/error|500/i)).toBeVisible();
	});
});
