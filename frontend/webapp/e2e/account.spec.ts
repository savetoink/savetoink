import { test, expect } from '@playwright/test';

test('unauthenticated /account page shows SharedKey login form', async ({ page }) => {
	await page.goto('/account');

	await expect(page.getByText('Account details')).toBeVisible();
	await expect(
		page.getByText('Enter your API key to access the article management system')
	).toBeVisible();

	const apiKeyInput = page.getByLabel('API Key');
	await expect(apiKeyInput).toHaveAttribute('type', 'password');
	await expect(apiKeyInput).toHaveAttribute('name', 'auth');
	await expect(apiKeyInput).toHaveAttribute('required', '');

	await expect(page.getByRole('button', { name: 'Login' })).toBeVisible();
});
