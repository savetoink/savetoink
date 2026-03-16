import { Page, expect } from '@playwright/test';

export async function waitForElement(page: Page, selector: string, timeout = 5000): Promise<void> {
	await page.waitForSelector(selector, { timeout, state: 'visible' });
}

export async function waitForElementHidden(page: Page, selector: string, timeout = 5000): Promise<void> {
	await page.waitForSelector(selector, { timeout, state: 'hidden' });
}

export async function waitForText(page: Page, selector: string, text: string, timeout = 5000): Promise<void> {
	await expect(page.locator(selector)).toHaveText(text, { timeout });
}

export async function waitForUrl(page: Page, url: string | RegExp, timeout = 5000): Promise<void> {
	await page.waitForURL(url, { timeout });
}

export async function takeScreenshot(page: Page, name: string): Promise<Buffer> {
	return await page.screenshot({ path: `screenshots/${name}.png` });
}

export async function waitForNetworkIdle(page: Page, timeout = 5000): Promise<void> {
	await page.waitForLoadState('networkidle', { timeout });
}

export function generateTestEmail(): string {
	return `test-${Date.now()}@example.com`;
}

export function generateTestUrl(): string {
	return `https://test-${Date.now()}.example.com/article`;
}
