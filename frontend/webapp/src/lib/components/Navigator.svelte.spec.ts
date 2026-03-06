import { page } from 'vitest/browser';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Navigator from './Navigator.svelte';

describe('Navigator.svelte', () => {
	it('should render only next link on first page with more pages', async () => {
		render(Navigator, { page: 1, has_more: true });

		const nextButton = page.getByRole('button', { name: 'Next' });
		await expect.element(nextButton).toBeInTheDocument();
	});

	it('should render only prev link on last page', async () => {
		render(Navigator, { page: 3, has_more: false });

		const prevButton = page.getByRole('button', { name: 'Previous' });
		await expect.element(prevButton).toBeInTheDocument();
	});

	it('should render both prev and next links on middle pages', async () => {
		render(Navigator, { page: 2, has_more: true });

		const prevButton = page.getByRole('button', { name: 'Previous' });
		await expect.element(prevButton).toBeInTheDocument();

		const nextButton = page.getByRole('button', { name: 'Next' });
		await expect.element(nextButton).toBeInTheDocument();
	});
});
