import type { Page } from '@playwright/test';
import { expect } from '@playwright/test';
import { BasePage } from './base.page';

export class AccountPage extends BasePage {
	constructor(page: Page) {
		super(page);
	}

	async navigateAccount() {
		await this.navigate('/account');
	}

	async expectOnAccountPage() {
		await expect(this.page).toHaveURL(/\/account/);
	}

	getAccountHeading() {
		return this.page.getByRole('heading', { name: 'Your account' });
	}

	async expectAccountHeading() {
		await expect(this.getAccountHeading()).toBeVisible();
	}

	getAuthForm() {
		return this.page.locator('form');
	}

	getDeviceDeliverySection() {
		return this.page.getByText(/Device delivery/);
	}
}
