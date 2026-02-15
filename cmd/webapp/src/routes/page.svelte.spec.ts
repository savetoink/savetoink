import { page } from 'vitest/browser';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Page from './+page.svelte';

describe('/+page.svelte', () => {
	it('should render article list', async () => {
		render(Page, {
			data: {
				articles: [],
				total: 0,
				page: 1,
				page_size: 10,
				has_more: false
			}
		});

		const heading = page.getByRole('heading', { name: 'My List' });
		await expect.element(heading).toBeInTheDocument();
	});

	it('should render next button when more articles exist', async () => {
		render(Page, {
			data: {
				articles: [
					{
						account: 'test-account',
						id: '1',
						url: 'https://example.com',
						createdAt: '2024-01-01T00:00:00Z',
						title: 'Test Article'
					}
				],
				total: 25,
				page: 1,
				page_size: 10,
				has_more: true
			}
		});

		const nextButton = page.getByRole('button', { name: 'Next' });
		await expect.element(nextButton).toBeInTheDocument();
	});

	it('should render prev button when not on first page', async () => {
		render(Page, {
			data: {
				articles: [
					{
						account: 'test-account',
						id: '1',
						url: 'https://example.com',
						createdAt: '2024-01-01T00:00:00Z',
						title: 'Test Article'
					}
				],
				total: 25,
				page: 2,
				page_size: 10,
				has_more: false
			}
		});

		const prevButton = page.getByRole('button', { name: 'Previous' });
		await expect.element(prevButton).toBeInTheDocument();
	});
});
