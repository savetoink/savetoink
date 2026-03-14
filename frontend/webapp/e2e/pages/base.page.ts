import type { Page } from '@playwright/test';

export class BasePage {
	constructor(readonly page: Page) {}

	async navigate(path: string = '/') {
		await this.page.goto(path);
		await this.page.waitForLoadState('networkidle');
	}

	async waitForNavigation() {
		await this.page.waitForLoadState('networkidle');
	}

	async getTitle() {
		return await this.page.title();
	}

	async getURL() {
		return this.page.url();
	}

	async screenshot(filename: string) {
		await this.page.screenshot({ path: `screenshots/${filename}` });
	}
}
