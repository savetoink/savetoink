import type { Page } from '@playwright/test';
import { expect } from '@playwright/test';
import { BasePage } from './base.page';

export class HomePage extends BasePage {
	constructor(page: Page) {
		super(page);
	}

	async expectOnHomePage() {
		await expect(this.page).toHaveURL(/\/$/);
	}

	async getNavLinks() {
		return await this.page.getByRole('navigation').getByRole('link').all();
	}

	async getNavLinkByText(text: string) {
		return this.page.getByRole('navigation').getByRole('link', { name: text });
	}

	async clickNavLink(text: string) {
		const link = await this.getNavLinkByText(text);
		await link.click();
		await this.waitForNavigation();
	}
}
