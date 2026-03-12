import { page } from 'vitest/browser';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Paginator from './Paginator.svelte';

describe('Paginator.svelte', () => {
	it('should render only next link on first page with more pages', async () => {
		render(Paginator, { page: 1, hasMore: true });

		const nextButton = page.getByRole('button', { name: 'Next' });
		await expect.element(nextButton).toBeInTheDocument();
	});

	it('should render only prev link on last page', async () => {
		render(Paginator, { page: 3, hasMore: false });

		const prevButton = page.getByRole('button', { name: 'Previous' });
		await expect.element(prevButton).toBeInTheDocument();
	});

	it('should render both prev and next links on middle pages', async () => {
		render(Paginator, { page: 2, hasMore: true });

		const prevButton = page.getByRole('button', { name: 'Previous' });
		await expect.element(prevButton).toBeInTheDocument();

		const nextButton = page.getByRole('button', { name: 'Next' });
		await expect.element(nextButton).toBeInTheDocument();
	});
});
