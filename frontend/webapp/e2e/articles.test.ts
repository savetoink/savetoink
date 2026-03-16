import { test, ArticlesPage } from './';

test.describe('Articles Page', () => {
	let articlesPage: ArticlesPage;

	test.beforeEach(async ({ page }) => {
		articlesPage = new ArticlesPage(page);
		await articlesPage.navigateToArticles();
	});

	test('should display articles list', async () => {
		await articlesPage.expectHasArticles();
	});

	test('should navigate up and down with keyboard', async () => {
		await articlesPage.navigateDown();
		await articlesPage.navigateUp();
	});

	test('should use keyboard shortcuts', async ({ keyboardHelper }) => {
		await keyboardHelper.pressKey('h');
		await keyboardHelper.pressKey('a');
	});
});
