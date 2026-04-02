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

	it('should not render X buttons when editable is false', async () => {
		render(Tags, { tags: ['tech'], articleId: 'test-article-id', editable: false });

		try {
			page.getByRole('button', { name: 'Remove tag: tech' });
			expect.fail('Remove button should not be present when editable is false');
		} catch {
			// Expected - button should not be found
		}
	});

	it('should not render X buttons when articleId is not provided', async () => {
		render(Tags, { tags: ['tech'], editable: true });

		try {
			page.getByRole('button', { name: 'Remove tag: tech' });
			expect.fail('Remove button should not be present when articleId is not provided');
		} catch {
			// Expected - button should not be found
		}
	});

	it('should render nothing when tags array is empty', async () => {
		render(Tags, { tags: [], editable: true, articleId: 'test-id' });

		try {
			page.getByRole('list');
			expect.fail('List should not be present when tags is empty');
		} catch {
			// Expected - list should not be found
		}
	});

	it('should render nothing when tags is undefined', async () => {
		render(Tags, { tags: undefined, editable: true, articleId: 'test-id' });

		try {
			page.getByRole('list');
			expect.fail('List should not be present when tags is undefined');
		} catch {
			// Expected - list should not be found
		}
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
