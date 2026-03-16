import { BasePage } from './BasePage';

export class AccountPage extends BasePage {
	async navigateToAccount(): Promise<void> {
		await this.navigate('/account');
		await this.waitForLoadState();
	}

	async expectAccountSectionVisible(): Promise<void> {
		await this.expectVisible('section:has-text("Your account")');
	}

	async isLoggedIn(): Promise<boolean> {
		return await this.isVisible('[data-test="device-delivery"]');
	}

	async expectLoggedIn(): Promise<void> {
		await this.expectVisible('[data-test="device-delivery"]');
	}

	async expectLoggedOut(): Promise<void> {
		await this.expectHidden('[data-test="device-delivery"]');
	}

	async navigateHome(): Promise<void> {
		await this.page.keyboard.press('h');
		await this.page.waitForURL('**/articles');
	}

	async navigateToArticles(): Promise<void> {
		await this.page.keyboard.press('a');
		await this.page.waitForURL('**/account');
	}
}
