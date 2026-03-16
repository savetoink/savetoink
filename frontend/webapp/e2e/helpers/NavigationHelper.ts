import { Page, expect } from '@playwright/test';

export class NavigationHelper {
	constructor(private readonly page: Page) {}

	async navigateTo(path: string): Promise<void> {
		await this.page.goto(path);
		await this.page.waitForLoadState('networkidle');
	}

	async expectUrl(expectedPath: string): Promise<void> {
		await expect(this.page).toHaveURL(expectedPath);
	}

	async expectUrlContains(fragment: string): Promise<void> {
		await expect(this.page).toHaveURL(new RegExp(fragment));
	}

	async getCurrentPath(): Promise<string> {
		const url = new URL(this.page.url());
		return url.pathname;
	}

	async reloadAndWait(): Promise<void> {
		await this.page.reload();
		await this.page.waitForLoadState('networkidle');
	}

	async goBack(): Promise<void> {
		await this.page.goBack();
		await this.page.waitForLoadState('networkidle');
	}

	async goForward(): Promise<void> {
		await this.page.goForward();
		await this.page.waitForLoadState('networkidle');
	}
}
