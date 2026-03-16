import { Page, expect } from '@playwright/test';

export class AssertionHelper {
	constructor(private readonly page: Page) {}

	async expectText(selector: string, text: string): Promise<void> {
		await expect(this.page.locator(selector)).toHaveText(text);
	}

	async expectTextContains(selector: string, text: string): Promise<void> {
		await expect(this.page.locator(selector)).toContainText(text);
	}

	async expectVisible(selector: string): Promise<void> {
		await expect(this.page.locator(selector)).toBeVisible();
	}

	async expectHidden(selector: string): Promise<void> {
		await expect(this.page.locator(selector)).toBeHidden();
	}

	async expectEnabled(selector: string): Promise<void> {
		await expect(this.page.locator(selector)).toBeEnabled();
	}

	async expectDisabled(selector: string): Promise<void> {
		await expect(this.page.locator(selector)).toBeDisabled();
	}

	async expectCount(selector: string, count: number): Promise<void> {
		await expect(this.page.locator(selector)).toHaveCount(count);
	}

	async expectAttribute(selector: string, attribute: string, value: string): Promise<void> {
		await expect(this.page.locator(selector)).toHaveAttribute(attribute, value);
	}

	async expectUrl(expectedPath: string): Promise<void> {
		await expect(this.page).toHaveURL(expectedPath);
	}

	async expectUrlContains(fragment: string): Promise<void> {
		await expect(this.page).toHaveURL(new RegExp(fragment));
	}
}
