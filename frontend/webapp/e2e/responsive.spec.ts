import { test, expect } from './fixtures';
import { TEST_URLS } from './utils/constants';
import { BREAKPOINTS } from './utils/constants';
import { setViewportSize } from './utils/helpers';

test.describe('Responsive Design', () => {
	test('should display correctly on mobile viewport', async ({ page }) => {
		await setViewportSize(page, BREAKPOINTS.MOBILE, 667);
		await page.goto(TEST_URLS.ACCOUNT);

		await expect(page.getByRole('navigation')).toBeVisible();
	});

	test('should display correctly on tablet viewport', async ({ page }) => {
		await setViewportSize(page, BREAKPOINTS.TABLET, 1024);
		await page.goto(TEST_URLS.ACCOUNT);

		await expect(page.getByRole('navigation')).toBeVisible();
	});

	test('should display correctly on desktop viewport', async ({ page }) => {
		await setViewportSize(page, BREAKPOINTS.DESKTOP, 800);
		await page.goto(TEST_URLS.ACCOUNT);

		await expect(page.getByRole('navigation')).toBeVisible();
	});

	test('should display correctly on large desktop viewport', async ({ page }) => {
		await setViewportSize(page, BREAKPOINTS.LARGE_DESKTOP, 1080);
		await page.goto(TEST_URLS.ACCOUNT);

		await expect(page.getByRole('navigation')).toBeVisible();
	});

	test('should handle orientation change', async ({ page }) => {
		await setViewportSize(page, 375, 667);
		await page.goto(TEST_URLS.ACCOUNT);

		const elementBefore = await page.getByRole('navigation').boundingBox();
		expect(elementBefore).toBeTruthy();

		await setViewportSize(page, 667, 375);
		const elementAfter = await page.getByRole('navigation').boundingBox();
		expect(elementAfter).toBeTruthy();
	});

	test('should maintain layout on window resize', async ({ page, pageFactory }) => {
		await page.goto(TEST_URLS.NEW);
		const newPage = pageFactory.newArticle();

		await setViewportSize(page, BREAKPOINTS.MOBILE, 667);
		await expect(newPage.getUrlInput()).toBeVisible();

		await setViewportSize(page, BREAKPOINTS.DESKTOP, 800);
		await expect(newPage.getUrlInput()).toBeVisible();
	});

	test('should have proper viewport meta tag', async ({ page }) => {
		await page.goto(TEST_URLS.HOME);

		const viewportMeta = page.locator('meta[name="viewport"]');
		await expect(viewportMeta).toHaveAttribute('content', /width=device-width/);
	});

	test('should scale form inputs correctly on different viewports', async ({
		page,
		pageFactory
	}) => {
		const newPage = pageFactory.newArticle();

		const viewports = [
			{ width: BREAKPOINTS.MOBILE, height: 667 },
			{ width: BREAKPOINTS.TABLET, height: 1024 },
			{ width: BREAKPOINTS.DESKTOP, height: 800 }
		];

		for (const vp of viewports) {
			await setViewportSize(page, vp.width, vp.height);
			await newPage.navigateNew();

			const urlInput = newPage.getUrlInput();
			await expect(urlInput).toBeVisible();

			const boundingBox = await urlInput.boundingBox();
			expect(boundingBox).toBeDefined();
			expect(boundingBox!.width).toBeGreaterThan(0);
		}
	});

	test('should handle text overflow on small screens', async ({ page }) => {
		await setViewportSize(page, 320, 568);
		await page.goto(TEST_URLS.ACCOUNT);

		const nav = page.getByRole('navigation');
		await expect(nav).toBeVisible();
	});

	test('should maintain readability on all viewport sizes', async ({ page }) => {
		const viewports = [
			{ width: BREAKPOINTS.MOBILE, height: 667 },
			{ width: BREAKPOINTS.DESKTOP, height: 800 }
		];

		for (const vp of viewports) {
			await setViewportSize(page, vp.width, vp.height);
			await page.goto(TEST_URLS.NEW);

			const heading = page.getByRole('heading', { level: 1 });
			await expect(heading).toBeVisible();

			const fontSize = await heading.evaluate((el) => {
				return window.getComputedStyle(el).fontSize;
			});
			expect(parseFloat(fontSize)).toBeGreaterThan(0);
		}
	});

	test('should handle touch interactions on mobile', async ({ page, pageFactory }) => {
		const isMobile =
			page.context().browser?.browserType().name() === 'webkit' ||
			page.context().browser?.browserType().name() === 'chromium';

		test.skip(!isMobile, 'Touch interactions only on mobile browsers');

		await setViewportSize(page, BREAKPOINTS.MOBILE, 667);
		const newPage = pageFactory.newArticle();
		await newPage.navigateNew();

		const urlInput = newPage.getUrlInput();
		await urlInput.tap();

		await expect(urlInput).toBeFocused();
	});
});
