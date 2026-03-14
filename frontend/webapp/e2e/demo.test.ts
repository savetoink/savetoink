import { test, expect } from './fixtures';
import { TEST_URLS, NAV_LINKS } from './utils/constants';

test.describe('Demo Tests', () => {
	test('account page has expected nav', async ({ page, pageFactory }) => {
		const accountPage = pageFactory.account();
		await accountPage.navigateAccount();
		await accountPage.expectAccountHeading();

		const nav = page.getByRole('navigation').filter({ hasText: NAV_LINKS.ACCOUNT });
		await expect(nav).toBeVisible();
	});

	test('basic navigation flow', async ({ page, pageFactory }) => {
		const homePage = pageFactory.home();

		await homePage.navigate(TEST_URLS.HOME);
		await expect(page).toHaveURL(TEST_URLS.HOME);

		await homePage.clickNavLink(NAV_LINKS.ACCOUNT);
		await expect(page).toHaveURL(TEST_URLS.ACCOUNT);
	});
});
