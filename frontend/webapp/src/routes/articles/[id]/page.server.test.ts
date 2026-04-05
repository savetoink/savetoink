import { describe, expect, it, beforeEach, vi } from 'vitest';
import type { RequestEvent } from '@sveltejs/kit';

/* eslint-disable @typescript-eslint/no-explicit-any */

const mockGetArticle = vi.fn();
const mockAddTags = vi.fn();

vi.mock('$lib/server/apiClient', async (importOriginal) => {
	const actual = await importOriginal<typeof import('$lib/server/apiClient')>();
	return {
		...actual,
		getArticle: mockGetArticle,
		addTags: mockAddTags
	};
});

describe('/articles/[id] - addTags action', () => {
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
		request: new Request('http://localhost/articles/test-id'),
		fetch: mockFetch,
		getClientAddress: mockGetClientAddress,
		url: new URL('http://localhost/articles/test-id'),
		params: { id: 'test-id' },
		route: { id: '/articles/[id]' },
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
		mockGetArticle.mockResolvedValue({
			id: 'test-id',
			tags: ['existing-tag']
		});
		mockAddTags.mockResolvedValue({
			tags: ['existing-tag', 'new-tag']
		});
	});

	describe('successful tag addition', () => {
		it('should add tags successfully', async () => {
			const formData = new FormData();
			formData.append('tags', 'tech, news, politics');

			const mockRequest = new Request('http://localhost/articles/test-id', {
				method: 'POST',
				body: formData
			});

			const eventWithForm = {
				...mockRequestEvent,
				request: mockRequest
			};

			const { actions } = await import('./+page.server');
			const result = await actions.addTags(eventWithForm as any);

			expect(mockGetArticle).toHaveBeenCalledWith(expect.any(Object), 'test-id');
			expect(mockAddTags).toHaveBeenCalledWith(expect.any(Object), 'test-id', [
				'tech',
				'news',
				'politics'
			]);
			expect(result).toEqual({
				tags: ['existing-tag', 'new-tag']
			});
		});

		it('should parse comma-separated tags and trim whitespace', async () => {
			const formData = new FormData();
			formData.append('tags', ' tech , news ,  politics  ');

			const mockRequest = new Request('http://localhost/articles/test-id', {
				method: 'POST',
				body: formData
			});

			const eventWithForm = {
				...mockRequestEvent,
				request: mockRequest
			};

			const { actions } = await import('./+page.server');
			await actions.addTags(eventWithForm as any);

			expect(mockAddTags).toHaveBeenCalledWith(expect.any(Object), 'test-id', [
				'tech',
				'news',
				'politics'
			]);
		});

		it('should add a single tag', async () => {
			const formData = new FormData();
			formData.append('tags', 'new-tag');

			const mockRequest = new Request('http://localhost/articles/test-id', {
				method: 'POST',
				body: formData
			});

			const eventWithForm = {
				...mockRequestEvent,
				request: mockRequest
			};

			const { actions } = await import('./+page.server');
			await actions.addTags(eventWithForm as any);

			expect(mockAddTags).toHaveBeenCalledWith(expect.any(Object), 'test-id', ['new-tag']);
		});

		it('should allow adding tags when exactly at max limit', async () => {
			mockGetArticle.mockResolvedValue({
				id: 'test-id',
				tags: ['tag1', 'tag2', 'tag3', 'tag4', 'tag5', 'tag6', 'tag7', 'tag8']
			});

			const formData = new FormData();
			formData.append('tags', 'tag9, tag10');

			const mockRequest = new Request('http://localhost/articles/test-id', {
				method: 'POST',
				body: formData
			});

			const eventWithForm = {
				...mockRequestEvent,
				request: mockRequest
			};

			const { actions } = await import('./+page.server');
			await actions.addTags(eventWithForm as any);

			expect(mockAddTags).toHaveBeenCalledWith(expect.any(Object), 'test-id', ['tag9', 'tag10']);
		});
	});

	describe('duplicate tags', () => {
		it('should send duplicate tags to backend for normalization', async () => {
			const formData = new FormData();
			formData.append('tags', 'tech, tech, news');

			const mockRequest = new Request('http://localhost/articles/test-id', {
				method: 'POST',
				body: formData
			});

			const eventWithForm = {
				...mockRequestEvent,
				request: mockRequest
			};

			const { actions } = await import('./+page.server');
			await actions.addTags(eventWithForm as any);

			// Backend should handle duplicate normalization
			expect(mockAddTags).toHaveBeenCalledWith(expect.any(Object), 'test-id', [
				'tech',
				'tech',
				'news'
			]);
		});
	});
});
