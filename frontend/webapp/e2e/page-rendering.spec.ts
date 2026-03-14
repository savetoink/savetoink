import { test, expect } from './fixtures';
import { TEST_URLS, HEADINGS, NAV_LINKS } from './utils/constants';

test.describe('Page Rendering', () => {
	test('should render home page with navigation', async ({ page }) => {
		await page.goto(TEST_URLS.HOME);

		await expect(page.getByRole('navigation')).toBeVisible();
	});

	test('should render account page with heading', async ({ page }) => {
		await page.goto(TEST_URLS.ACCOUNT);

		await expect(page.getByRole('heading', { name: HEADINGS.ACCOUNT })).toBeVisible();
	});

	test('should render new article page with form', async ({ page }) => {
		await page.goto(TEST_URLS.NEW);

		await expect(page.getByRole('heading', { name: HEADINGS.ADD_ARTICLE })).toBeVisible();
		await expect(page.getByRole('form')).toBeVisible();
	});

	test('should render articles page', async ({ page }) => {
		await page.goto(TEST_URLS.ARTICLES);

		await expect(page.getByRole('main')).toBeVisible();
	});

	test('should display navigation links correctly', async ({ page }) => {
		await page.goto(TEST_URLS.HOME);

		const nav = page.getByRole('navigation');
		await expect(nav).toBeVisible();

		const accountLink = nav.getByRole('link', { name: NAV_LINKS.ACCOUNT });
		await expect(accountLink).toBeVisible();
	});

	test('should render footer on all pages', async ({ page }) => {
		await page.goto(TEST_URLS.HOME);
		await expect(page.locator('footer')).toBeVisible();

		await page.goto(TEST_URLS.ACCOUNT);
		await expect(page.locator('footer')).toBeVisible();

		await page.goto(TEST_URLS.NEW);
		await expect(page.locator('footer')).toBeVisible();
	});

	test('should have proper page title', async ({ page }) => {
		await page.goto(TEST_URLS.HOME);
		const title = await page.title();
		expect(title.length).toBeGreaterThan(0);
	});

	test('should render meta tags correctly', async ({ page }) => {
		await page.goto(TEST_URLS.HOME);

		const favicon = page.locator('link[rel="icon"]');
		await expect(favicon).toHaveCount(1);
	});

	test('should have proper document language', async ({ page }) => {
		await page.goto(TEST_URLS.HOME);

		const lang = await page.locator('html').getAttribute('lang');
		expect(lang).toBeTruthy();
	});

	test('should render favorites page', async ({ page }) => {
		await page.goto(TEST_URLS.ARTICLES_FAVORITES);

		await expect(page.getByRole('main')).toBeVisible();
	});
});
