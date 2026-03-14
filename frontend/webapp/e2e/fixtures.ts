import { test as base } from '@playwright/test';
import { PageFactory } from './pages/factory';

type MyFixtures = {
	pageFactory: PageFactory;
};

export const test = base.extend<MyFixtures>({
	pageFactory: async ({ page }, use) => {
		await use(new PageFactory(page));
	}
});

export { expect } from '@playwright/test';
