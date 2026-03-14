# E2E Tests

This directory contains end-to-end tests for the webapp using Playwright.

## Structure

```
e2e/
├── pages/              # Page Object Model classes
│   ├── base.page.ts    # Base page with common methods
│   ├── home.page.ts    # Home page object
│   ├── articles.page.ts # Articles page object
│   ├── new-article.page.ts # New article page object
│   ├── account.page.ts # Account page object
│   └── factory.ts      # Page factory for creating page objects
├── utils/              # Test utilities
│   ├── constants.ts    # Test constants (URLs, selectors, etc.)
│   └── helpers.ts      # Helper functions
├── fixtures.ts         # Playwright fixtures
├── navigation.spec.ts  # Navigation tests
├── page-rendering.spec.ts # Page rendering tests
├── form-validation.spec.ts # Form validation tests
├── responsive.spec.ts  # Responsive design tests
└── demo.test.ts        # Demo tests
```

## Page Object Model

The test suite uses the Page Object Model pattern for better maintainability:

- **BasePage**: Common methods for all pages (navigation, waiting, screenshots)
- **HomePage**: Home page specific interactions
- **ArticlesPage**: Articles list page interactions
- **NewArticlePage**: New article form interactions
- **AccountPage**: Account page interactions
- **PageFactory**: Factory for creating page objects

## Running Tests

Run all e2e tests:

```bash
bun run test:e2e
```

Run specific test file:

```bash
bun run test:e2e navigation.spec.ts
```

Run tests in headed mode:

```bash
bun run test:e2e --headed
```

Run tests with UI:

```bash
bun run test:e2e --ui
```

## Test Coverage

Current e2e tests cover:

- Navigation between pages
- Page rendering and layout
- Form validation (new article form)
- Responsive design across different viewports
- Basic accessibility checks

## Adding New Tests

1. Create a new test file following the naming convention `*.spec.ts` or `*.test.ts`
2. Use the custom fixtures from `fixtures.ts` for access to page factory
3. Use page objects from the `pages/` directory for interactions
4. Import constants and helpers from `utils/`

Example:

```typescript
import { test, expect } from './fixtures';
import { TEST_URLS } from './utils/constants';

test.describe('My Feature', () => {
	test('should do something', async ({ page, pageFactory }) => {
		const homePage = pageFactory.home();
		await homePage.navigate(TEST_URLS.HOME);

		await expect(page.getByText('Hello')).toBeVisible();
	});
});
```

## Debugging

To debug a failing test:

1. Run with `--debug` flag: `bun run test:e2e --debug`
2. Use VS Code Playwright extension for step-by-step debugging
3. Check screenshots in `test-results/screenshots/` for failures
4. View HTML report: `bun run test:e2e --reporter=html` then open `playwright-report/index.html`
