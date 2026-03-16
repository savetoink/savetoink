import { Page, expect } from '@playwright/test';

export class FormHelper {
	constructor(private readonly page: Page) {}

	async fillForm(formData: Record<string, string>): Promise<void> {
		for (const [name, value] of Object.entries(formData)) {
			await this.page.fill(`[name="${name}"]`, value);
		}
	}

	async submitForm(selector = 'form'): Promise<void> {
		await this.page.click(`${selector} button[type="submit"]`);
	}

	async expectFormVisible(selector = 'form'): Promise<void> {
		await expect(this.page.locator(selector)).toBeVisible();
	}

	async expectFormHidden(selector = 'form'): Promise<void> {
		await expect(this.page.locator(selector)).toBeHidden();
	}

	async getInputValue(name: string): Promise<string | null> {
		const input = this.page.locator(`[name="${name}"]`);
		return await input.inputValue();
	}

	async clearInput(name: string): Promise<void> {
		await this.page.fill(`[name="${name}"]`, '');
	}

	async selectOption(name: string, value: string): Promise<void> {
		await this.page.selectOption(`[name="${name}"]`, value);
	}

	async checkCheckbox(name: string): Promise<void> {
		await this.page.check(`[name="${name}"]`);
	}

	async uncheckCheckbox(name: string): Promise<void> {
		await this.page.uncheck(`[name="${name}"]`);
	}
}
