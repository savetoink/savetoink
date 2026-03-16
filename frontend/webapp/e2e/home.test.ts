import { test, HomePage } from './';

test.describe('Home Page', () => {
	test.beforeEach(async ({ page }) => {
		const homePage = new HomePage(page);
		await homePage.navigateToHome();
	});

	test('should navigate to articles', async ({ navigationHelper }) => {
		await navigationHelper.navigateTo('/articles');
		await navigationHelper.expectUrl('/articles');
	});

	test('should navigate to account', async ({ navigationHelper }) => {
		await navigationHelper.navigateTo('/account');
		await navigationHelper.expectUrl('/account');
	});

	test('should navigate to new article', async ({ navigationHelper }) => {
		await navigationHelper.navigateTo('/new');
		await navigationHelper.expectUrl('/new');
	});
});
