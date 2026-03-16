import { BasePage } from './BasePage';

export class NewArticlePage extends BasePage {
	async navigateToNewArticle(): Promise<void> {
		await this.navigate('/new');
		await this.waitForLoadState();
	}

	async expectFormVisible(): Promise<void> {
		await this.expectVisible('form');
	}

	async fillUrl(url: string): Promise<void> {
		await this.fillInput('input[name="url"]', url);
	}

	async submitForm(): Promise<void> {
		await this.page.click('button[type="submit"]');
	}

	async createArticle(url: string): Promise<void> {
		await this.fillUrl(url);
		await this.submitForm();
	}

	async expectSuccess(): Promise<void> {
		await this.page.waitForURL('**/articles');
	}
}
