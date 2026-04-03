import { page } from 'vitest/browser';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Tags from './Tags.svelte';

describe('Tags.svelte', () => {
	it('should render tags as links when not editable', async () => {
		render(Tags, { tags: ['tech', 'news', 'politics'], editable: false });

		const tagLinks = page.getByRole('link', { name: /#tech/i });
		await expect.element(tagLinks).toBeInTheDocument();

		// Check that remove button is not present
		try {
			page.getByRole('button', { name: 'Remove tag: tech' });
			expect.fail('Remove button should not be present when not editable');
		} catch {
			// Expected - button should not be found
		}
	});

	it('should render X buttons when editable and articleId is provided', async () => {
		render(Tags, { tags: ['tech', 'news'], articleId: 'test-article-id', editable: true });

		const techRemoveButton = page.getByRole('button', { name: 'Remove tag: tech' });
		await expect.element(techRemoveButton).toBeInTheDocument();

		const newsRemoveButton = page.getByRole('button', { name: 'Remove tag: news' });
		await expect.element(newsRemoveButton).toBeInTheDocument();
	});

	it('should render proper aria-label for remove button', async () => {
		render(Tags, { tags: ['tech'], articleId: 'test-id', editable: true });

		const removeButton = page.getByRole('button', { name: 'Remove tag: tech' });
		await expect.element(removeButton).toBeInTheDocument();
	});

	it('should render tag link inside remove button when editable', async () => {
		render(Tags, { tags: ['tech'], articleId: 'test-id', editable: true });

		const tagLink = page.getByRole('link', { name: /#tech/i });
		await expect.element(tagLink).toBeInTheDocument();

		// Verify link is inside the remove button
		const removeButton = page.getByRole('button', { name: 'Remove tag: tech' });
		expect(removeButton).toContainElement(tagLink);
	});
});
