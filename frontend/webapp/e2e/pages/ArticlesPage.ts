import { BasePage } from './BasePage';

export class ArticlesPage extends BasePage {
	async navigateToArticles(): Promise<void> {
		await this.navigate('/articles');
		await this.waitForLoadState();
	}

	async getArticleCount(): Promise<number> {
		return await this.page.locator('ul > li').count();
	}

	async selectArticle(index: number): Promise<void> {
		const _article = this.page.locator('ul > li').nth(index);
		await _article.click();
	}

	async selectArticleById(id: string): Promise<void> {
		const article = this.page.locator(`ul > li[data-article-id="${id}"]`);
		await article.click();
	}

	async openArticle(index: number): Promise<void> {
		const _article = this.page.locator('ul > li').nth(index);
		await _article.dblclick();
	}

	async expectNoArticles(): Promise<void> {
		await this.expectVisible('p:has-text("No articles found")');
	}

	async expectHasArticles(): Promise<void> {
		const count = await this.getArticleCount();
		expect(count).toBeGreaterThan(0);
	}

	async getArticleTitle(index: number): Promise<string | null> {
		const article = this.page.locator('ul > li').nth(index);
		return await article.textContent();
	}

	async clickFavorite(): Promise<void> {
		await this.page.keyboard.press('k');
		await this.page.keyboard.press('ArrowDown');
		await this.page.keyboard.press('f');
	}

	async clickDelete(): Promise<void> {
		await this.page.keyboard.press('k');
		await this.page.keyboard.press('ArrowDown');
		await this.page.keyboard.press('d');
	}

	async navigateDown(): Promise<void> {
		await this.page.keyboard.press('ArrowDown');
	}

	async navigateUp(): Promise<void> {
		await this.page.keyboard.press('ArrowUp');
	}

	async openSelectedArticle(): Promise<void> {
		await this.page.keyboard.press('Enter');
	}

	async clickKeyboardShortcut(key: string): Promise<void> {
		await this.page.keyboard.press(key);
	}
}
