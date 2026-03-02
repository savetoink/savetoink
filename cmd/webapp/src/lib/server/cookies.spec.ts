import { describe, expect, it, beforeEach, vi } from 'vitest';
import type { Cookies } from '@sveltejs/kit';

const mockSet = vi.fn();
const mockDelete = vi.fn();
const mockGet = vi.fn();
const mockGetAll = vi.fn();
const mockSerialize = vi.fn();

const mockCookies = {
	set: mockSet,
	delete: mockDelete,
	get: mockGet,
	getAll: mockGetAll,
	serialize: mockSerialize
} as unknown as Cookies;

vi.mock('@sveltejs/kit', async () => {
	const actual = await vi.importActual('@sveltejs/kit');
	return {
		...actual
	};
});

describe('cookies', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	describe('setJwtCookie', () => {
		it('should set cookie with default options', async () => {
			const { setAuthCookie: setJwtCookie } = await import('./cookies');
			const testToken = 'test-token-123';

			setJwtCookie(mockCookies, testToken);

			expect(mockSet).toHaveBeenCalledWith('auth', testToken, {
				path: '/',
				httpOnly: true,
				secure: false,
				sameSite: 'lax',
				maxAge: 60 * 60 * 24 * 365
			});
		});

		it('should trim token when trim option is true', async () => {
			const { setAuthCookie: setJwtCookie } = await import('./cookies');
			const testToken = '  test-token-123  ';

			setJwtCookie(mockCookies, testToken, { trim: true });

			expect(mockSet).toHaveBeenCalledWith('auth', 'test-token-123', expect.any(Object));
		});

		it('should use custom maxAge', async () => {
			const { setAuthCookie: setJwtCookie } = await import('./cookies');
			const testToken = 'test-token-123';
			const customMaxAge = 60 * 60 * 24;

			setJwtCookie(mockCookies, testToken, { maxAge: customMaxAge });

			expect(mockSet).toHaveBeenCalledWith(
				'auth',
				testToken,
				expect.objectContaining({
					maxAge: customMaxAge
				})
			);
		});

		it('should use custom secure option', async () => {
			const { setAuthCookie: setJwtCookie } = await import('./cookies');
			const testToken = 'test-token-123';

			setJwtCookie(mockCookies, testToken, { secure: true });

			expect(mockSet).toHaveBeenCalledWith(
				'auth',
				testToken,
				expect.objectContaining({
					secure: true
				})
			);
		});
	});

	describe('deleteJwtCookie', () => {
		it('should delete cookie', async () => {
			const { deleteAuthCookie: deleteJwtCookie } = await import('./cookies');

			deleteJwtCookie(mockCookies);

			expect(mockDelete).toHaveBeenCalledWith('auth', { path: '/' });
		});
	});

	describe('AUTH_KEY', () => {
		it('should export AUTH_KEY constant', async () => {
			const { AUTH_KEY } = await import('./cookies');

			expect(AUTH_KEY).toBe('auth');
		});
	});
});
