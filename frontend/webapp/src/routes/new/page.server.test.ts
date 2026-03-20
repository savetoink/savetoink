import { describe, expect, it, beforeEach, vi } from 'vitest';
import { redirect } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';

/* eslint-disable @typescript-eslint/no-explicit-any */

vi.mock('@sveltejs/kit', async () => {
	const actual = await vi.importActual('@sveltejs/kit');
	return {
		...actual,
		redirect: vi.fn()
	};
});

const mockCreateArticle = vi.fn();

vi.mock('$lib/server/apiClient', async () => {
	return {
		createArticle: mockCreateArticle
	};
});

describe('/new route', () => {
	const mockLocals = {
		auth: 'test-auth-token',
		user: {
			account: 'test-account',
			email: 'test@example.com',
			device_email: 'test@kindle.com',
			auto_send: true
		}
	};

	const mockGetClientAddress = vi.fn(() => '127.0.0.1');
	const mockFetch = vi.fn();

	const mockRequestEvent = {
		locals: mockLocals,
		request: new Request('http://localhost/new'),
		fetch: mockFetch,
		getClientAddress: mockGetClientAddress,
		url: new URL('http://localhost/new'),
		params: {},
		route: { id: '/new' },
		isDataRequest: false,
		isSubRequest: false,
		isRemoteRequest: false,
		cookies: {} as Record<string, unknown>,
		platform: {} as Record<string, unknown>,
		setHeaders: vi.fn(),
		tracing: vi.fn(),
		parent: async () => ({}),
		depends: vi.fn(),
		untrack: vi.fn((fn) => fn())
	} as unknown as RequestEvent;

	beforeEach(() => {
		vi.clearAllMocks();
	});

	describe('load function', () => {
		it('should return user data when no URL parameter is provided', async () => {
			const { load } = await import('./+page.server');

			const result = await load(mockRequestEvent as any);

			expect(result).toEqual({
				user: mockLocals.user,
				incomingUrl: null
			});
			expect(mockCreateArticle).not.toHaveBeenCalled();
			expect(redirect).not.toHaveBeenCalled();
		});

		it('should return incomingUrl when url parameter is provided', async () => {
			const { load } = await import('./+page.server');

			const urlWithParam = new URL('http://localhost/new?url=https://example.com/article');
			const eventWithUrl = { ...mockRequestEvent, url: urlWithParam };

			const result = await load(eventWithUrl as any);

			expect(result).toEqual({
				user: mockLocals.user,
				incomingUrl: 'https://example.com/article'
			});
			expect(mockCreateArticle).not.toHaveBeenCalled();
			expect(redirect).not.toHaveBeenCalled();
		});

		it('should return incomingUrl when text parameter is provided', async () => {
			const { load } = await import('./+page.server');

			const urlWithParam = new URL('http://localhost/new?text=https://example.com/article');
			const eventWithUrl = { ...mockRequestEvent, url: urlWithParam };

			const result = await load(eventWithUrl as any);

			expect(result).toEqual({
				user: mockLocals.user,
				incomingUrl: 'https://example.com/article'
			});
			expect(mockCreateArticle).not.toHaveBeenCalled();
			expect(redirect).not.toHaveBeenCalled();
		});

		it('should prefer url parameter over text parameter', async () => {
			const { load } = await import('./+page.server');

			const urlWithParams = new URL(
				'http://localhost/new?url=https://first.com&text=https://second.com'
			);
			const eventWithUrl = { ...mockRequestEvent, url: urlWithParams };

			const result = await load(eventWithUrl as any);

			expect(result).toEqual({
				user: mockLocals.user,
				incomingUrl: 'https://first.com'
			});
			expect(mockCreateArticle).not.toHaveBeenCalled();
		});
	});

	describe('new action', () => {
		beforeEach(() => {
			mockCreateArticle.mockResolvedValue({
				id: 'test-article-id',
				title: 'Test Article Title',
				url: 'https://example.com/article'
			});
		});

		it('should create article and return success data when form is submitted', async () => {
			const { actions } = await import('./+page.server');

			const formData = new FormData();
			formData.append('url', 'https://example.com/article');
			formData.append('sendToDevice', 'on');

			const mockRequest = new Request('http://localhost/new', {
				method: 'POST',
				body: formData
			});

			const eventWithForm = {
				...mockRequestEvent,
				request: mockRequest
			};

			const result = await actions.new(eventWithForm as any);

			expect(mockCreateArticle).toHaveBeenCalledWith(
				{
					locals: mockLocals,
					request: mockRequest,
					fetch: mockFetch,
					getClientAddress: mockGetClientAddress
				},
				'https://example.com/article',
				true
			);
			expect(result).toEqual({
				success: true,
				article: {
					id: 'test-article-id',
					title: 'Test Article Title',
					url: 'https://example.com/article'
				}
			});
			expect(redirect).not.toHaveBeenCalled();
		});

		it('should create article without sending to device when checkbox not checked', async () => {
			const { actions } = await import('./+page.server');

			const formData = new FormData();
			formData.append('url', 'https://example.com/article');

			const mockRequest = new Request('http://localhost/new', {
				method: 'POST',
				body: formData
			});

			const eventWithForm = {
				...mockRequestEvent,
				request: mockRequest
			};

			const result = await actions.new(eventWithForm as any);

			expect(mockCreateArticle).toHaveBeenCalledWith(
				{
					locals: mockLocals,
					request: mockRequest,
					fetch: mockFetch,
					getClientAddress: mockGetClientAddress
				},
				'https://example.com/article',
				false
			);
			expect(result).toEqual({
				success: true,
				article: {
					id: 'test-article-id',
					title: 'Test Article Title',
					url: 'https://example.com/article'
				}
			});
		});

		it('should return error when URL is missing', async () => {
			const { actions } = await import('./+page.server');

			const formData = new FormData();
			formData.append('url', '');

			const mockRequest = new Request('http://localhost/new', {
				method: 'POST',
				body: formData
			});

			const eventWithForm = {
				...mockRequestEvent,
				request: mockRequest
			};

			const result = await actions.new(eventWithForm as any);

			expect((result as { status: number }).status).toBe(400);
			expect((result as { data: { error: string } }).data).toEqual({ error: 'URL is required' });
			expect(mockCreateArticle).not.toHaveBeenCalled();
		});

		it('should return error when createArticle fails', async () => {
			const { actions } = await import('./+page.server');

			mockCreateArticle.mockRejectedValue(new Error('Failed to fetch article'));

			const formData = new FormData();
			formData.append('url', 'https://example.com/article');

			const mockRequest = new Request('http://localhost/new', {
				method: 'POST',
				body: formData
			});

			const eventWithForm = {
				...mockRequestEvent,
				request: mockRequest
			};

			const result = await actions.new(eventWithForm as any);

			expect((result as { status: number }).status).toBe(500);
			expect((result as { data: { error: string } }).data).toEqual({
				error: 'Failed to create article: Failed to fetch article'
			});
		});

		it('should return error with correct status when createArticle throws HTTP error', async () => {
			const { actions } = await import('./+page.server');

			const httpError = new Error('Invalid URL');
			Object.assign(httpError, { status: 400 });
			mockCreateArticle.mockRejectedValue(httpError);

			const formData = new FormData();
			formData.append('url', 'https://example.com/article');

			const mockRequest = new Request('http://localhost/new', {
				method: 'POST',
				body: formData
			});

			const eventWithForm = {
				...mockRequestEvent,
				request: mockRequest
			};

			const result = await actions.new(eventWithForm as any);

			expect((result as { status: number }).status).toBe(400);
			expect((result as { data: { error: string } }).data).toEqual({
				error: 'Failed to create article: Invalid URL'
			});
		});
	});
});
