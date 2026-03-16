export const MOCK_ARTICLES = [
	{
		id: '1',
		title: 'Test Article 1',
		url: 'https://example.com/article1',
		created_at: '2024-01-01T00:00:00Z',
		updated_at: '2024-01-01T00:00:00Z',
		favorite: false
	},
	{
		id: '2',
		title: 'Test Article 2',
		url: 'https://example.com/article2',
		created_at: '2024-01-02T00:00:00Z',
		updated_at: '2024-01-02T00:00:00Z',
		favorite: true
	}
];

export const MOCK_URLS = [
	'https://example.com/test-article',
	'https://github.com/test/repo',
	'https://example.org/blog/post'
];

export const KEYBOARD_BINDINGS = {
	UP: 'ArrowUp',
	DOWN: 'ArrowDown',
	LEFT: 'ArrowLeft',
	RIGHT: 'ArrowRight',
	ENTER: 'Enter',
	ESCAPE: 'Escape',
	FAVORITE: 'f',
	DELETE: 'd',
	SEND: 's',
	NEW: 'n',
	HOME: 'h',
	ACCOUNT: 'a',
	NEXT: 'j',
	PREVIOUS: 'k'
};

export const NAVIGATION_ROUTES = {
	HOME: '/',
	ARTICLES: '/articles',
	ACCOUNT: '/account',
	NEW_ARTICLE: '/new'
};
