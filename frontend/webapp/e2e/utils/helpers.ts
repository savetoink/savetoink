import type { Page } from '@playwright/test';
import { expect } from '@playwright/test';

export async function expectNavigation(
	page: Page,
	expectedUrl: string | RegExp,
	timeout: number = 5000
) {
	await expect(async () => {
		const url = page.url();
		if (typeof expectedUrl === 'string') {
			expect(url).toContain(expectedUrl);
		} else {
			expect(url).toMatch(expectedUrl);
		}
	}).toPass({ timeout });
}

export async function expectElementVisible(page: Page, selector: string, timeout: number = 5000) {
	await expect(async () => {
		const element = page.locator(selector);
		await expect(element).toBeVisible();
	}).toPass({ timeout });
}

export async function expectElementText(
	page: Page,
	selector: string,
	expectedText: string,
	timeout: number = 5000
) {
	await expect(async () => {
		const element = page.locator(selector);
		await expect(element).toHaveText(expectedText);
	}).toPass({ timeout });
}

export async function expectElementCount(
	page: Page,
	selector: string,
	expectedCount: number,
	timeout: number = 5000
) {
	await expect(async () => {
		const elements = page.locator(selector);
		const count = await elements.count();
		expect(count).toBe(expectedCount);
	}).toPass({ timeout });
}

export async function waitForNavigationComplete(page: Page) {
	await page.waitForLoadState('networkidle');
	await page.waitForLoadState('domcontentloaded');
}

export async function setViewportSize(page: Page, width: number, height: number = 800) {
	await page.setViewportSize({ width, height });
}

export async function takeScreenshot(page: Page, name: string) {
	await page.screenshot({
		path: `test-results/screenshots/${name}.png`,
		fullPage: true
	});
}

export function getRandomString(length: number = 10) {
	return Math.random()
		.toString(36)
		.substring(2, 2 + length);
}

export function getValidTestUrl() {
	return `https://example.com/article-${getRandomString()}`;
}

export function getInvalidUrl() {
	return 'not-a-valid-url';
}
