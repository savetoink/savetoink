import type { Page } from '@playwright/test';
import { expect } from '@playwright/test';
import { BasePage } from './base.page';

export class NewArticlePage extends BasePage {
	constructor(page: Page) {
		super(page);
	}

	async navigateNew() {
		await this.navigate('/new');
	}

	async expectOnNewPage() {
		await expect(this.page).toHaveURL(/\/new/);
	}

	async expectPageTitle(text: string) {
		await expect(this.page.getByRole('heading', { level: 1 })).toHaveText(text);
	}

	getUrlInput() {
		return this.page.getByLabel('URL');
	}

	getSubmitButton() {
		return this.page.getByRole('button', { name: 'Add' });
	}

	getSendToDeviceCheckbox() {
		return this.page.getByRole('checkbox', { name: /Send to device/ });
	}

	async fillUrl(url: string) {
		await this.getUrlInput().fill(url);
	}

	async clickSendToDevice() {
		await this.getSendToDeviceCheckbox().check();
	}

	async uncheckSendToDevice() {
		await this.getSendToDeviceCheckbox().uncheck();
	}

	async submitForm() {
		await this.getSubmitButton().click();
	}
}
