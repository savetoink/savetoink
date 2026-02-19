import { describe, expect, it } from 'vitest';

describe('isAuthenticatedPath', () => {
	it('should return false for /login path', async () => {
		const { isAuthenticatedPath } = await import('./consts');
		expect(isAuthenticatedPath('/login')).toBe(false);
	});

	it('should return false for /auth/callback path', async () => {
		const { isAuthenticatedPath } = await import('./consts');
		expect(isAuthenticatedPath('/auth/callback')).toBe(false);
	});

	it('should return true for / path', async () => {
		const { isAuthenticatedPath } = await import('./consts');
		expect(isAuthenticatedPath('/')).toBe(true);
	});

	it('should return true for /dashboard path', async () => {
		const { isAuthenticatedPath } = await import('./consts');
		expect(isAuthenticatedPath('/dashboard')).toBe(true);
	});

	it('should return true for /api/some-endpoint path', async () => {
		const { isAuthenticatedPath } = await import('./consts');
		expect(isAuthenticatedPath('/api/some-endpoint')).toBe(true);
	});

	it('should return true for paths with query parameters', async () => {
		const { isAuthenticatedPath } = await import('./consts');
		expect(isAuthenticatedPath('/dashboard?foo=bar')).toBe(true);
	});

	it('should return true for paths with trailing slash', async () => {
		const { isAuthenticatedPath } = await import('./consts');
		expect(isAuthenticatedPath('/dashboard/')).toBe(true);
	});
});
