import { BasePage } from './BasePage';

export class ArticleDetailPage extends BasePage {
	async navigateToArticle(id: string): Promise<void> {
		await this.navigate(`/articles/${id}`);
		await this.waitForLoadState();
	}

	async expectArticleVisible(): Promise<void> {
		await this.expectVisible('[data-test="article-detail"]');
	}

	async navigateBack(): Promise<void> {
		await this.page.keyboard.press('Escape');
		await this.page.waitForURL('**/articles');
	}

	async expectBackNavigation(): Promise<void> {
		await this.page.waitForURL('**/articles');
	}

	async toggleFavorite(): Promise<void> {
		await this.page.keyboard.press('f');
	}

	async deleteArticle(): Promise<void> {
		await this.page.keyboard.press('d');
	}

	async sendToDevice(): Promise<void> {
		await this.page.keyboard.press('s');
	}
}
