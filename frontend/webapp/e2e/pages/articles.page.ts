import type { Page } from '@playwright/test';
import { expect } from '@playwright/test';
import { BasePage } from './base.page';

export class ArticlesPage extends BasePage {
	constructor(page: Page) {
		super(page);
	}

	async navigateArticles() {
		await this.navigate('/articles');
	}

	async navigateFavorites() {
		await this.navigate('/articles?favorite=true');
	}

	async expectOnArticlesPage() {
		await expect(this.page).toHaveURL(/\/articles/);
	}

	async getArticleItems() {
		return await this.page.locator('ul > li').all();
	}

	async getArticleByIndex(index: number) {
		const articles = await this.getArticleItems();
		return articles[index];
	}

	getEmptyState() {
		return this.page.getByText('No articles found');
	}

	async expectEmptyState() {
		await expect(this.getEmptyState()).toBeVisible();
	}

	getNextButton() {
		return this.page.getByRole('button', { name: 'Next' });
	}

	getPrevButton() {
		return this.page.getByRole('button', { name: 'Previous' });
	}
}
