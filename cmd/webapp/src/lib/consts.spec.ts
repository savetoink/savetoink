import { describe, expect, it } from 'vitest';

describe('isAuthenticatedPath', () => {
	it('should return false for /account path', async () => {
		const { isAuthenticatedPath } = await import('./consts');
		expect(isAuthenticatedPath('/account')).toBe(false);
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

	it('should return true for /account/ with trailing slash', async () => {
		const { isAuthenticatedPath } = await import('./consts');
		expect(isAuthenticatedPath('/account/')).toBe(true);
	});

	it('should return true for /auth/callback with query params', async () => {
		const { isAuthenticatedPath } = await import('./consts');
		expect(isAuthenticatedPath('/auth/callback?code=abc123')).toBe(true);
	});

	it('should return true for /account with extra segments', async () => {
		const { isAuthenticatedPath } = await import('./consts');
		expect(isAuthenticatedPath('/account/extra')).toBe(true);
	});

	it('should be case sensitive', async () => {
		const { isAuthenticatedPath } = await import('./consts');
		expect(isAuthenticatedPath('/Login')).toBe(true);
		expect(isAuthenticatedPath('/LOGIN')).toBe(true);
	});

	it('should export Auth0 constant', async () => {
		const { Auth0 } = await import('./consts');
		expect(Auth0).toBe('auth0');
	});

	it('should export SharedKey constant', async () => {
		const { SharedKey } = await import('./consts');
		expect(SharedKey).toBe('sharedKey');
	});
});
