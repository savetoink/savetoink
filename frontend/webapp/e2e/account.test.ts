import { test, expect, AccountPage } from './';

test.describe('Account Page', () => {
	test.beforeEach(async ({ page }) => {
		const accountPage = new AccountPage(page);
		await accountPage.navigateToAccount();
	});

	test('should display account section', async ({ page }) => {
		const accountPage = new AccountPage(page);
		await accountPage.expectAccountSectionVisible();
	});

	test('should have navigation available', async ({ page, keyboardHelper }) => {
		await keyboardHelper.pressKey('h');
	});
});
