import { describe, expect, it, beforeEach, vi } from 'vitest';
import type { Cookies } from '@sveltejs/kit';

vi.mock('$env/dynamic/private', () => ({
	env: {
		COOKIE_SECRET: ''
	}
}));

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

			expect(mockDelete).toHaveBeenCalledWith('auth', { path: '/', secure: false });
		});
	});

	describe('setRefreshCookie', () => {
		it('should set cookie with provided maxAge', async () => {
			const { setRefreshCookie } = await import('./cookies');
			const testToken = 'test-refresh-token';
			const maxAge = 60 * 60 * 24 * 30;

			setRefreshCookie(mockCookies, testToken, maxAge);

			expect(mockSet).toHaveBeenCalledWith('refresh', testToken, {
				path: '/',
				httpOnly: true,
				secure: false,
				sameSite: 'lax',
				maxAge
			});
		});
	});

	describe('deleteRefreshCookie', () => {
		it('should delete refresh cookie', async () => {
			const { deleteRefreshCookie } = await import('./cookies');

			deleteRefreshCookie(mockCookies);

			expect(mockDelete).toHaveBeenCalledWith('refresh', { path: '/', secure: false });
		});
	});

	describe('setTokenCookies', () => {
		it('should set auth and refresh cookies from token response', async () => {
			const { setTokenCookies } = await import('./cookies');

			setTokenCookies(mockCookies, {
				access_token: 'test-access',
				expires_in: 10,
				refresh_token: 'test-refresh',
				refresh_expires_in: 2592000
			});

			expect(mockSet).toHaveBeenCalledWith('auth', 'test-access', {
				path: '/',
				httpOnly: true,
				secure: false,
				sameSite: 'lax',
				maxAge: 10
			});
			expect(mockSet).toHaveBeenCalledWith('refresh', 'test-refresh', {
				path: '/',
				httpOnly: true,
				secure: false,
				sameSite: 'lax',
				maxAge: 2592000
			});
		});

		it('should skip refresh cookie when refresh_token is missing', async () => {
			const { setTokenCookies } = await import('./cookies');

			setTokenCookies(mockCookies, {
				access_token: 'test-access',
				expires_in: 10
			});

			expect(mockSet).toHaveBeenCalledTimes(1);
			expect(mockSet).toHaveBeenCalledWith('auth', 'test-access', expect.any(Object));
		});

		it('should use default maxAge when refresh_expires_in is not provided', async () => {
			const { setTokenCookies } = await import('./cookies');

			setTokenCookies(mockCookies, {
				access_token: 'test-access',
				expires_in: 10,
				refresh_token: 'test-refresh'
			});

			expect(mockSet).toHaveBeenCalledWith('refresh', 'test-refresh', {
				path: '/',
				httpOnly: true,
				secure: false,
				sameSite: 'lax',
				maxAge: 60 * 60 * 24 * 30
			});
		});
	});

	describe('AUTH_KEY', () => {
		it('should export AUTH_KEY constant', async () => {
			const { AUTH_KEY } = await import('./cookies');

			expect(AUTH_KEY).toBe('auth');
		});
	});

	describe('User cookie encoding and decoding', () => {
		it('should encode and decode user data', async () => {
			const { getUserCookie, setUserCookie } = await import('./cookies');
			const userData = {
				account: 'test-account',
				email: 'test@example.com',
				device_email: 'test@kindle.com',
				auto_send: true
			};

			mockGet.mockReturnValue(null);

			await setUserCookie(mockCookies, userData);

			expect(mockSet).toHaveBeenCalledWith(
				'profile',
				expect.stringContaining(JSON.stringify(userData)),
				expect.any(Object)
			);

			const cookieValue = mockSet.mock.calls[0][1];
			mockGet.mockReturnValue(cookieValue);

			const decoded = await getUserCookie(mockCookies);

			expect(decoded).toEqual(userData);
		});

		it('should return undefined for missing cookie', async () => {
			const { getUserCookie } = await import('./cookies');

			mockGet.mockReturnValue(null);

			const result = await getUserCookie(mockCookies);

			expect(result).toBeUndefined();
		});

		it('should return null for invalid signature', async () => {
			const { getUserCookie, setUserCookie } = await import('./cookies');
			const userData = {
				account: 'test-account',
				email: 'test@example.com',
				device_email: 'test@kindle.com',
				auto_send: true
			};

			mockGet.mockReturnValue(null);
			await setUserCookie(mockCookies, userData);

			const cookieValue = mockSet.mock.calls[0][1];
			const tamperedValue = cookieValue.slice(0, -10) + '0000000000';
			mockGet.mockReturnValue(tamperedValue);

			const decoded = await getUserCookie(mockCookies);

			expect(decoded).toBeUndefined();
		});
	});

	describe('deleteUserCookie', () => {
		it('should delete user cookie', async () => {
			const { deleteUserCookie } = await import('./cookies');

			await deleteUserCookie(mockCookies);

			expect(mockDelete).toHaveBeenCalledWith('profile', { path: '/', secure: false });
		});
	});

	describe('redirect_to cookie functions', () => {
		describe('setRedirectToCookie', () => {
			it('should set cookie with correct options', async () => {
				const { setRedirectToCookie } = await import('./cookies');
				const testUrl = '/articles?page=2';

				setRedirectToCookie(mockCookies, testUrl);

				expect(mockSet).toHaveBeenCalledWith('redirect_to', testUrl, {
					path: '/',
					httpOnly: true,
					secure: false,
					sameSite: 'lax',
					maxAge: 300 // 5 minutes
				});
			});

			it('should set cookie with URL including pathname and query params', async () => {
				const { setRedirectToCookie } = await import('./cookies');
				const testUrl = '/articles?page=2&filter=recent';

				setRedirectToCookie(mockCookies, testUrl);

				expect(mockSet).toHaveBeenCalledWith(
					'redirect_to',
					'/articles?page=2&filter=recent',
					expect.any(Object)
				);
			});

			it('should set cookie with simple pathname', async () => {
				const { setRedirectToCookie } = await import('./cookies');
				const testUrl = '/articles';

				setRedirectToCookie(mockCookies, testUrl);

				expect(mockSet).toHaveBeenCalledWith('redirect_to', '/articles', expect.any(Object));
			});
		});

		describe('REDIRECT_TO_KEY', () => {
			it('should export REDIRECT_TO_KEY constant', async () => {
				const { REDIRECT_TO_KEY } = await import('./cookies');

				expect(REDIRECT_TO_KEY).toBe('redirect_to');
			});
		});

		describe('getValidRedirectUrl', () => {
			beforeEach(() => {
				vi.clearAllMocks();
			});

			it('should return valid redirect URL and delete cookie', async () => {
				const { getValidRedirectUrl } = await import('./cookies');
				const testUrl = '/articles?page=2';

				mockGet.mockReturnValue(testUrl);

				const result = getValidRedirectUrl(mockCookies);

				expect(result).toBe(testUrl);
				expect(mockDelete).toHaveBeenCalledWith('redirect_to', { path: '/', secure: false });
			});

			it('should return undefined when no cookie exists', async () => {
				const { getValidRedirectUrl } = await import('./cookies');

				mockGet.mockReturnValue(null);

				const result = getValidRedirectUrl(mockCookies);

				expect(result).toBeUndefined();
				expect(mockDelete).not.toHaveBeenCalled();
			});

			it('should return undefined and delete cookie for invalid URL', async () => {
				const { getValidRedirectUrl } = await import('./cookies');

				mockGet.mockReturnValue('https://evil.com');

				const result = getValidRedirectUrl(mockCookies);

				expect(result).toBeUndefined();
				expect(mockDelete).toHaveBeenCalledWith('redirect_to', { path: '/', secure: false });
			});

			it('should return undefined and delete cookie for protocol-relative URL', async () => {
				const { getValidRedirectUrl } = await import('./cookies');

				mockGet.mockReturnValue('//evil.com');

				const result = getValidRedirectUrl(mockCookies);

				expect(result).toBeUndefined();
				expect(mockDelete).toHaveBeenCalledWith('redirect_to', { path: '/', secure: false });
			});

			it('should return valid URL and delete cookie for simple path', async () => {
				const { getValidRedirectUrl } = await import('./cookies');
				const testUrl = '/articles';

				mockGet.mockReturnValue(testUrl);

				const result = getValidRedirectUrl(mockCookies);

				expect(result).toBe(testUrl);
				expect(mockDelete).toHaveBeenCalledWith('redirect_to', { path: '/', secure: false });
			});

			it('should return valid URL for root path', async () => {
				const { getValidRedirectUrl } = await import('./cookies');
				const testUrl = '/';

				mockGet.mockReturnValue(testUrl);

				const result = getValidRedirectUrl(mockCookies);

				expect(result).toBe('/');
				expect(mockDelete).toHaveBeenCalledWith('redirect_to', { path: '/', secure: false });
			});

			it('should return valid URL with multiple query parameters', async () => {
				const { getValidRedirectUrl } = await import('./cookies');
				const testUrl = '/articles?page=2&filter=recent&sort=desc';

				mockGet.mockReturnValue(testUrl);

				const result = getValidRedirectUrl(mockCookies);

				expect(result).toBe(testUrl);
				expect(mockDelete).toHaveBeenCalledWith('redirect_to', { path: '/', secure: false });
			});

			it('should return undefined and delete cookie for URL without leading slash', async () => {
				const { getValidRedirectUrl } = await import('./cookies');

				mockGet.mockReturnValue('articles');

				const result = getValidRedirectUrl(mockCookies);

				expect(result).toBeUndefined();
				expect(mockDelete).toHaveBeenCalledWith('redirect_to', { path: '/', secure: false });
			});

			it('should return undefined and delete cookie for URL with javascript protocol', async () => {
				const { getValidRedirectUrl } = await import('./cookies');

				mockGet.mockReturnValue('javascript:alert(1)');

				const result = getValidRedirectUrl(mockCookies);

				expect(result).toBeUndefined();
				expect(mockDelete).toHaveBeenCalledWith('redirect_to', { path: '/', secure: false });
			});

			it('should return undefined and delete cookie for URL with data protocol', async () => {
				const { getValidRedirectUrl } = await import('./cookies');

				mockGet.mockReturnValue('data:text/html,<script>alert(1)</script>');

				const result = getValidRedirectUrl(mockCookies);

				expect(result).toBeUndefined();
				expect(mockDelete).toHaveBeenCalledWith('redirect_to', { path: '/', secure: false });
			});
		});
	});
});
