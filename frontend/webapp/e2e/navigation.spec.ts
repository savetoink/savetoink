import { test, expect } from './fixtures';
import { TEST_URLS, NAV_LINKS } from './utils/constants';
import { waitForNavigationComplete } from './utils/helpers';

test.describe('Navigation', () => {
	test.beforeEach(async ({ page }) => {
		await page.goto(TEST_URLS.HOME);
	});

	test('should navigate to account page', async ({ page, pageFactory }) => {
		const homePage = pageFactory.home();
		await homePage.clickNavLink(NAV_LINKS.ACCOUNT);

		expect(page.url()).toContain(TEST_URLS.ACCOUNT);
		await expect(page.getByRole('heading', { name: 'Your account' })).toBeVisible();
	});

	test('should navigate between pages and update URL', async ({ page, pageFactory }) => {
		const homePage = pageFactory.home();
		const accountPage = pageFactory.account();

		await homePage.clickNavLink(NAV_LINKS.ACCOUNT);
		await accountPage.expectOnAccountPage();

		await pageFactory.home().navigate(TEST_URLS.HOME);
		expect(page.url()).toContain(TEST_URLS.HOME);
	});

	test('should maintain navigation state when reloading', async ({ page }) => {
		await page.goto(TEST_URLS.ACCOUNT);
		const urlBefore = page.url();

		await page.reload();
		await waitForNavigationComplete(page);

		expect(page.url()).toBe(urlBefore);
	});

	test('should handle navigation history correctly', async ({ page }) => {
		await page.goto(TEST_URLS.ACCOUNT);
		await page.goto(TEST_URLS.HOME);
		await page.goto(TEST_URLS.ACCOUNT);

		await page.goBack();
		expect(page.url()).toContain(TEST_URLS.HOME);

		await page.goForward();
		expect(page.url()).toContain(TEST_URLS.ACCOUNT);
	});

	test('should show progress indicator during navigation', async ({ page }) => {
		await page.goto(TEST_URLS.ACCOUNT);
		const link = page.getByRole('link', { name: NAV_LINKS.ACCOUNT });

		await link.click();
		const progress = page.locator('progress');
		await expect(progress).toBeVisible();

		await waitForNavigationComplete(page);
	});

	test('should handle 404 for non-existent routes', async ({ page }) => {
		await page.goto('/non-existent-route');

		expect(page.status()).toBe(404);
	});

	test('should navigate to favorites via nav link', async ({ page, pageFactory }) => {
		const homePage = pageFactory.home();
		await homePage.clickNavLink(NAV_LINKS.FAVORITES);

		expect(page.url()).toContain('favorite=true');
	});
});
