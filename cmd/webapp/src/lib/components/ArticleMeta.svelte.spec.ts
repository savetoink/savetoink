import { page } from 'vitest/browser';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import ArticleMeta from './ArticleMeta.svelte';
import type { Article as ArticleType } from '$lib/server/types';

describe('ArticleMeta.svelte', () => {
	describe('mode="card"', () => {
		it('should render card with (almost) all fields', async () => {
			const article: ArticleType = {
				account: 'test-account',
				id: '1',
				url: 'https://example.com/article',
				createdAt: '2024-01-01T00:00:00Z',
				title: 'Test Article',
				content: 'Test content',
				author: 'Test Author',
				siteName: 'Example Site',
				sourceDomain: 'example.com',
				excerpt: 'Test excerpt',
				imageUrl: 'https://example.com/image.jpg',
				wordCount: 500,
				readingTimeMinutes: 2,
				deliveryStatus: 'pending',
				publishedAt: '2023-12-01T00:00:00Z',
				deliveredFrom: 'test@example.com',
				deliveredTo: 'kindle@kindle.com'
			};

			render(ArticleMeta, { article, mode: 'card' });

			const titleLink = page.getByRole('link', { name: 'Test Article' });
			await expect.element(titleLink).toBeInTheDocument();

			const image = page.getByRole('img');
			await expect.element(image).toBeInTheDocument();

			const excerpt = page.getByText('Test excerpt');
			await expect.element(excerpt).toBeInTheDocument();

			const originalLink = page.getByRole('link', { name: 'https://example.com/article' });
			await expect.element(originalLink).toHaveAttribute('target', '_blank');
			await expect.element(originalLink).toHaveAttribute('rel', 'external');

			const author = page.getByText('by Test Author');
			await expect.element(author).toBeInTheDocument();

			const published = page.getByText('published:');
			await expect.element(published).toBeInTheDocument();

			const readingTime = page.getByText('2 min read');
			await expect.element(readingTime).toBeInTheDocument();

			const status = page.getByText('status: pending');
			await expect.element(status).toBeInTheDocument();
		});

		it('should render card with minimal fields', async () => {
			const article: ArticleType = {
				account: 'test-account',
				id: '1',
				url: 'https://example.com/article',
				createdAt: '2024-01-01T00:00:00Z'
			};

			render(ArticleMeta, { article, mode: 'card' });

			const heading = page.getByRole('heading');
			await expect.element(heading).toBeInTheDocument();

			const originalLink = page.getByRole('link', { name: 'https://example.com/article' }).nth(1);
			await expect.element(originalLink).toHaveAttribute('target', '_blank');
			await expect.element(originalLink).toHaveAttribute('rel', 'external');
		});

		it('should render card error', async () => {
			const article: ArticleType = {
				account: 'test-account',
				id: '1',
				url: 'https://example.com/article',
				createdAt: '2024-01-01T00:00:00Z',
				error: 'Failed to fetch content'
			};

			render(ArticleMeta, { article, mode: 'card' });

			const error = page.getByText('error: Failed to fetch content');
			await expect.element(error).toBeInTheDocument();
		});

		it('should render article controls', async () => {
			const article: ArticleType = {
				account: 'test-account',
				id: 'test-id-123',
				url: 'https://example.com/article',
				createdAt: '2024-01-01T00:00:00Z',
				title: 'Test Article'
			};

			render(ArticleMeta, { article, mode: 'card' });

			const deleteButton = page.getByRole('button', { name: 'Delete' });
			await expect.element(deleteButton).toBeInTheDocument();
		});
	});

	describe('mode="header"', () => {
		it('should render header with all fields', async () => {
			const article: ArticleType = {
				account: 'test-account',
				id: '1',
				url: 'https://example.com/article',
				createdAt: '2024-01-01T00:00:00Z',
				title: 'Test Article',
				author: 'Test Author',
				siteName: 'Example Site',
				sourceDomain: 'example.com',
				wordCount: 500,
				readingTimeMinutes: 2,
				deliveryStatus: 'delivered',
				publishedAt: '2023-12-01T00:00:00Z',
				deliveredBy: 'system',
				deliveredEmailUUID: 'abc-123-def'
			};

			render(ArticleMeta, { article, mode: 'header' });

			const title = page.getByRole('heading', { level: 1, name: 'Test Article' });
			await expect.element(title).toBeInTheDocument();

			const author = page.getByText('by Test Author');
			await expect.element(author).toBeInTheDocument();

			const source = page.getByText('source: Example Site');
			await expect.element(source).toBeInTheDocument();

			const domain = page.getByText('domain: example.com');
			await expect.element(domain).toBeInTheDocument();

			const wordCount = page.getByText('500 words');
			await expect.element(wordCount).toBeInTheDocument();

			const readingTime = page.getByText('2 min read');
			await expect.element(readingTime).toBeInTheDocument();

			const status = page.getByText('status: delivered');
			await expect.element(status).toBeInTheDocument();

			const deliveredBy = page.getByText('delivered by: system');
			await expect.element(deliveredBy).toBeInTheDocument();

			const emailUUID = page.getByText('email id: abc-123-def');
			await expect.element(emailUUID).toBeInTheDocument();
		});

		it('should render header with minimal fields', async () => {
			const article: ArticleType = {
				account: 'test-account',
				id: '1',
				url: 'https://example.com/article',
				createdAt: '2024-01-01T00:00:00Z'
			};

			render(ArticleMeta, { article, mode: 'header' });

			const heading = page.getByRole('heading', { level: 1, name: 'article' });
			await expect.element(heading).toBeInTheDocument();
		});

		it('should render header error', async () => {
			const article: ArticleType = {
				account: 'test-account',
				id: '1',
				url: 'https://example.com/article',
				createdAt: '2024-01-01T00:00:00Z',
				error: 'Failed to process'
			};

			render(ArticleMeta, { article, mode: 'header' });

			const error = page.getByText('error: Failed to process');
			await expect.element(error).toBeInTheDocument();
		});

		it('should use header element', async () => {
			const article: ArticleType = {
				account: 'test-account',
				id: '1',
				url: 'https://example.com/article',
				createdAt: '2024-01-01T00:00:00Z',
				title: 'Test Article'
			};

			const { container } = render(ArticleMeta, { article, mode: 'header' });

			const header = container.querySelector('header');
			expect(header).toBeDefined();
		});
	});
});
