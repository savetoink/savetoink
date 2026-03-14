import type { Page } from '@playwright/test';
import { HomePage } from './home.page';
import { ArticlesPage } from './articles.page';
import { NewArticlePage } from './new-article.page';
import { AccountPage } from './account.page';

export class PageFactory {
	constructor(private page: Page) {}

	home() {
		return new HomePage(this.page);
	}

	articles() {
		return new ArticlesPage(this.page);
	}

	newArticle() {
		return new NewArticlePage(this.page);
	}

	account() {
		return new AccountPage(this.page);
	}
}
