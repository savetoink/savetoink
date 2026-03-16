import { test as base } from '@playwright/test';
import { KeyboardHelper } from '../helpers/KeyboardHelper';
import { NavigationHelper } from '../helpers/NavigationHelper';
import { FormHelper } from '../helpers/FormHelper';
import { AssertionHelper } from '../helpers/AssertionHelper';

type TestFixtures = {
	keyboardHelper: KeyboardHelper;
	navigationHelper: NavigationHelper;
	formHelper: FormHelper;
	assertionHelper: AssertionHelper;
};

export const test = base.extend<TestFixtures>({
	keyboardHelper: async ({ page }, use) => {
		const helper = new KeyboardHelper(page);
		await use(helper);
	},
	navigationHelper: async ({ page }, use) => {
		const helper = new NavigationHelper(page);
		await use(helper);
	},
	formHelper: async ({ page }, use) => {
		const helper = new FormHelper(page);
		await use(helper);
	},
	assertionHelper: async ({ page }, use) => {
		const helper = new AssertionHelper(page);
		await use(helper);
	}
});

export const expect = base.expect;
