import { BasePage } from './BasePage';

export class HomePage extends BasePage {
	async navigateToHome(): Promise<void> {
		await this.navigate('/');
		await this.waitForLoadState();
	}

	async navigateToArticles(): Promise<void> {
		await this.navigate('/articles');
		await this.waitForLoadState();
	}

	async navigateToAccount(): Promise<void> {
		await this.navigate('/account');
		await this.waitForLoadState();
	}

	async navigateToNewArticle(): Promise<void> {
		await this.navigate('/new');
		await this.waitForLoadState();
	}

	async expectNavigation(): Promise<void> {
		await this.page.waitForURL('**');
	}
}
