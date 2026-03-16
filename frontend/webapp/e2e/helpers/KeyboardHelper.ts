import { Page } from '@playwright/test';

export class KeyboardHelper {
	constructor(private readonly page: Page) {}

	async pressKey(key: string): Promise<void> {
		await this.page.keyboard.press(key);
	}

	async pressKeys(keys: string[]): Promise<void> {
		for (const key of keys) {
			await this.pressKey(key);
		}
	}

	async typeText(text: string, delay = 50): Promise<void> {
		await this.page.keyboard.type(text, { delay });
	}

	async selectAll(): Promise<void> {
		if (process.platform === 'darwin') {
			await this.pressKey('Meta+A');
		} else {
			await this.pressKey('Control+A');
		}
	}

	async copy(): Promise<void> {
		if (process.platform === 'darwin') {
			await this.pressKey('Meta+C');
		} else {
			await this.pressKey('Control+C');
		}
	}

	async paste(): Promise<void> {
		if (process.platform === 'darwin') {
			await this.pressKey('Meta+V');
		} else {
			await this.pressKey('Control+V');
		}
	}
}
