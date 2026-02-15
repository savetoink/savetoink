import { page } from 'vitest/browser';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Nav from './Nav.svelte';

describe('Nav.svelte', () => {
	it('should render navigation links', async () => {
		render(Nav);

		const myListLink = page.getByRole('link', { name: 'My List' });
		await expect.element(myListLink).toBeInTheDocument();
		await expect.element(myListLink).toHaveAttribute('href', '/');

		const saveLink = page.getByRole('link', { name: 'New' });
		await expect.element(saveLink).toBeInTheDocument();
		await expect.element(saveLink).toHaveAttribute('href', '/new');

		const settingsLink = page.getByRole('link', { name: 'Settings' });
		await expect.element(settingsLink).toBeInTheDocument();
		await expect.element(settingsLink).toHaveAttribute('href', '/settings');
	});
});
