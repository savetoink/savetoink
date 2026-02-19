import type { Cookies } from '@sveltejs/kit';

const JWT_KEY = 'jwt';

export interface SetJwtCookieOptions {
	trim?: boolean;
	maxAge?: number;
	secure?: boolean;
}

export function setJwtCookie(cookies: Cookies, token: string, options?: SetJwtCookieOptions) {
	const {
		trim = false,
		maxAge = 60 * 60 * 24 * 365,
		secure = import.meta.env.PROD
	} = options ?? {};

	const value = trim ? token.trim() : token;
	cookies.set(JWT_KEY, value, {
		path: '/',
		httpOnly: true,
		secure,
		sameSite: 'lax',
		maxAge
	});
}

export function deleteJwtCookie(cookies: Cookies) {
	cookies.delete(JWT_KEY, { path: '/' });
}

export function getJwtCookie(cookies: Cookies) {
	return cookies.get(JWT_KEY);
}

export { JWT_KEY };
