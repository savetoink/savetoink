import { test, expect } from './fixtures';
import { TEST_URLS, BUTTONS, FORM_LABELS } from './utils/constants';
import { getValidTestUrl, getInvalidUrl } from './utils/helpers';

test.describe('Form Validation - New Article', () => {
	test.beforeEach(async ({ page }) => {
		await page.goto(TEST_URLS.NEW);
	});

	test('should require URL field', async ({ pageFactory }) => {
		const newPage = pageFactory.newArticle();

		await newPage.submitForm();

		const urlInput = newPage.getUrlInput();
		await expect(urlInput).toBeFocused();
	});

	test('should validate URL format', async ({ pageFactory }) => {
		const newPage = pageFactory.newArticle();

		await newPage.fillUrl(getInvalidUrl());
		await newPage.submitForm();

		const urlInput = newPage.getUrlInput();
		await expect(urlInput).toHaveAttribute('type', 'url');
	});

	test('should accept valid URL', async ({ pageFactory }) => {
		const newPage = pageFactory.newArticle();

		await newPage.fillUrl(getValidTestUrl());
		const urlInput = newPage.getUrlInput();
		await expect(urlInput).toHaveValue(/https?:\/\//);
	});

	test('should have proper form labels', async ({ page }) => {
		await expect(page.getByLabel(FORM_LABELS.URL)).toBeVisible();
	});

	test('should have submit button with correct text', async ({ page }) => {
		await expect(page.getByRole('button', { name: BUTTONS.ADD })).toBeVisible();
	});

	test('should autofocus URL input on page load', async ({ pageFactory }) => {
		const newPage = pageFactory.newArticle();
		await newPage.navigateNew();

		const urlInput = newPage.getUrlInput();
		await expect(urlInput).toBeFocused();
	});

	test('should have placeholder text for URL input', async ({ page }) => {
		const urlInput = page.getByLabel(FORM_LABELS.URL);
		await expect(urlInput).toHaveAttribute('placeholder', /https?:\/\/example\.com/);
	});

	test('should handle send to device checkbox', async ({ pageFactory }) => {
		const newPage = pageFactory.newArticle();

		const checkbox = newPage.getSendToDeviceCheckbox();
		const isInitiallyChecked = await checkbox.isChecked();

		await newPage.clickSendToDevice();
		await expect(checkbox).toBeChecked();

		if (!isInitiallyChecked) {
			await newPage.uncheckSendToDevice();
			await expect(checkbox).not.toBeChecked();
		}
	});

	test('should prevent form submission with empty URL', async ({ page, pageFactory }) => {
		const newPage = pageFactory.newArticle();

		await newPage.submitForm();

		expect(page.url()).toContain(TEST_URLS.NEW);
	});

	test('should show form action attribute', async ({ page }) => {
		const form = page.locator('form[method="POST"]');
		await expect(form).toHaveAttribute('action', '?/new');
	});

	test('should have autocomplete attribute on URL input', async ({ page }) => {
		const urlInput = page.getByLabel(FORM_LABELS.URL);
		await expect(urlInput).toHaveAttribute('autocomplete', 'url');
	});

	test('should have proper fieldset structure', async ({ page }) => {
		await expect(page.locator('fieldset')).toBeVisible();
	});

	test('should have proper heading structure', async ({ page }) => {
		await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
		await expect(page.getByRole('heading', { level: 2 })).toBeVisible();
	});
});
