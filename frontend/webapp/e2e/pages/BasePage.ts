import { expect, Page } from '@playwright/test';

export class BasePage {
	constructor(protected readonly page: Page) {}

	async navigate(path: string): Promise<void> {
		await this.page.goto(path);
	}

	async waitForLoadState(state: 'load' | 'domcontentloaded' | 'networkidle' = 'load'): Promise<void> {
		await this.page.waitForLoadState(state);
	}

	async reload(): Promise<void> {
		await this.page.reload();
	}

	async getTitle(): Promise<string> {
		return await this.page.title();
	}

	async getElementText(selector: string): Promise<string | null> {
		const element = this.page.locator(selector).first();
		return await element.textContent();
	}

	async clickElement(selector: string): Promise<void> {
		await this.page.click(selector);
	}

	async fillInput(selector: string, value: string): Promise<void> {
		await this.page.fill(selector, value);
	}

	async isVisible(selector: string): Promise<boolean> {
		return await this.page.locator(selector).isVisible();
	}

	async isHidden(selector: string): Promise<boolean> {
		return await this.page.locator(selector).isHidden();
	}

	async expectVisible(selector: string): Promise<void> {
		await expect(this.page.locator(selector)).toBeVisible();
	}

	async expectHidden(selector: string): Promise<void> {
		await expect(this.page.locator(selector)).toBeHidden();
	}

	async expectText(selector: string, text: string): Promise<void> {
		await expect(this.page.locator(selector)).toHaveText(text);
	}

	async waitForElement(selector: string, timeout?: number): Promise<void> {
		await this.page.waitForSelector(selector, { timeout });
	}

	async waitForElementHidden(selector: string, timeout?: number): Promise<void> {
		await this.page.waitForSelector(selector, { state: 'hidden', timeout });
	}
}
