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
				user: mockLocals.user
			});
			expect(mockCreateArticle).not.toHaveBeenCalled();
			expect(redirect).not.toHaveBeenCalled();
		});

		it('should create article and redirect when url parameter is provided', async () => {
			const { load } = await import('./+page.server');

			const urlWithParam = new URL('http://localhost/new?url=https://example.com/article');
			const eventWithUrl = { ...mockRequestEvent, url: urlWithParam };

			await load(eventWithUrl as any);

			expect(mockCreateArticle).toHaveBeenCalledWith(
				{
					request: { headers: expect.any(Headers) },
					fetch: mockFetch,
					getClientAddress: expect.any(Function),
					locals: { auth: mockLocals.auth }
				},
				'https://example.com/article',
				true
			);
			expect(redirect).toHaveBeenCalledWith(303, '/articles');
		});

		it('should create article with text parameter', async () => {
			const { load } = await import('./+page.server');

			const urlWithParam = new URL('http://localhost/new?text=https://example.com/article');
			const eventWithUrl = { ...mockRequestEvent, url: urlWithParam };

			await load(eventWithUrl as any);

			expect(mockCreateArticle).toHaveBeenCalledWith(
				{
					request: { headers: expect.any(Headers) },
					fetch: mockFetch,
					getClientAddress: expect.any(Function),
					locals: { auth: mockLocals.auth }
				},
				'https://example.com/article',
				true
			);
			expect(redirect).toHaveBeenCalledWith(303, '/articles');
		});

		it('should respect auto_send preference when auto-creating article', async () => {
			const { load } = await import('./+page.server');

			const mockLocalsNoAutoSend = {
				auth: 'test-auth-token',
				user: {
					account: 'test-account',
					email: 'test@example.com',
					device_email: 'test@kindle.com',
					auto_send: false
				}
			};

			const urlWithParam = new URL('http://localhost/new?url=https://example.com/article');
			const eventWithUrl = { ...mockRequestEvent, url: urlWithParam, locals: mockLocalsNoAutoSend };

			await load(eventWithUrl as any);

			expect(mockCreateArticle).toHaveBeenCalledWith(
				{
					request: { headers: expect.any(Headers) },
					fetch: mockFetch,
					getClientAddress: expect.any(Function),
					locals: { auth: mockLocalsNoAutoSend.auth }
				},
				'https://example.com/article',
				false
			);
		});

		it('should default sendToDevice to false when user is undefined', async () => {
			const { load } = await import('./+page.server');

			const mockLocalsNoUser = {
				auth: 'test-auth-token',
				user: undefined
			};

			const urlWithParam = new URL('http://localhost/new?url=https://example.com/article');
			const eventWithUrl = { ...mockRequestEvent, url: urlWithParam, locals: mockLocalsNoUser };

			await load(eventWithUrl as any);

			expect(mockCreateArticle).toHaveBeenCalledWith(
				{
					request: { headers: expect.any(Headers) },
					fetch: mockFetch,
					getClientAddress: expect.any(Function),
					locals: { auth: mockLocalsNoUser.auth }
				},
				'https://example.com/article',
				false
			);
		});

		it('should prefer url parameter over text parameter', async () => {
			const { load } = await import('./+page.server');

			const urlWithParams = new URL(
				'http://localhost/new?url=https://first.com&text=https://second.com'
			);
			const eventWithUrl = { ...mockRequestEvent, url: urlWithParams };

			await load(eventWithUrl as any);

			expect(mockCreateArticle).toHaveBeenCalledWith(
				{
					request: { headers: expect.any(Headers) },
					fetch: mockFetch,
					getClientAddress: expect.any(Function),
					locals: { auth: mockLocals.auth }
				},
				'https://first.com',
				true
			);
		});
	});

	describe('new action', () => {
		it('should create article and redirect when form is submitted', async () => {
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

			await actions.new(eventWithForm as any);

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
			expect(redirect).toHaveBeenCalledWith(303, '/articles');
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

			await actions.new(eventWithForm as any);

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
		});
	});
});
