export const TEST_URLS = {
	HOME: '/',
	ARTICLES: '/articles',
	ARTICLES_FAVORITES: '/articles?favorite=true',
	NEW: '/new',
	ACCOUNT: '/account'
} as const;

export const NAV_LINKS = {
	ADD_NEW: 'Add new',
	ARTICLES: 'Articles',
	FAVORITES: 'Favorites',
	ACCOUNT: 'My account'
} as const;

export const HEADINGS = {
	HOME: 'Save to Ink',
	ADD_ARTICLE: 'Add Article',
	ACCOUNT: 'Your account'
} as const;

export const FORM_LABELS = {
	URL: 'URL',
	SEND_TO_DEVICE: 'Send to device'
} as const;

export const BUTTONS = {
	ADD: 'Add',
	SUBMIT: 'Submit',
	NEXT: 'Next',
	PREVIOUS: 'Previous',
	DELETE: 'Delete',
	SEND: 'Send',
	FAVORITE: 'Favorite'
} as const;

export const MESSAGES = {
	NO_ARTICLES: 'No articles found',
	INVALID_URL: 'Please enter a valid URL',
	REQUIRED_FIELD: 'This field is required'
} as const;

export const BREAKPOINTS = {
	MOBILE: 375,
	TABLET: 768,
	DESKTOP: 1280,
	LARGE_DESKTOP: 1920
} as const;
