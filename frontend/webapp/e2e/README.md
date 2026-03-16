# E2E Tests

End-to-end tests for the webapp using Playwright.

## Structure

- `pages/` - Page object models for each major page
- `helpers/` - Reusable helper classes for common operations
- `fixtures.ts` - Custom Playwright test fixtures
- `constants.ts` - Test constants and mock data
- `utils.ts` - Utility functions for tests
- `*.test.ts` - Actual test files

## Page Objects

- `BasePage` - Base page with common methods
- `HomePage` - Home page navigation
- `ArticlesPage` - Articles list page
- `AccountPage` - Account page
- `NewArticlePage` - New article creation
- `ArticleDetailPage` - Article detail view

## Helpers

- `KeyboardHelper` - Keyboard interaction helper
- `NavigationHelper` - Navigation utilities
- `FormHelper` - Form interaction utilities
- `AssertionHelper` - Common assertions

## Running Tests

```bash
# Run all e2e tests
bun run test:e2e

# Run tests in headed mode
bun run test:e2e --headed

# Run specific test file
bun run test:e2e articles.test.ts

# Run tests in debug mode
bun run test:e2e --debug

# View test report
bun run test:e2e --reporter=html
```

## Writing Tests

Use the page objects and helpers for consistency:

```typescript
import { test, expect, ArticlesPage } from './';

test('should display articles', async ({ page }) => {
	const articlesPage = new ArticlesPage(page);
	await articlesPage.navigateToArticles();
	await articlesPage.expectHasArticles();
});
```
