import { createHmac, timingSafeEqual } from 'crypto';
import type { Cookies } from '@sveltejs/kit';
import type { User } from '$lib/types';

const AUTH_KEY = 'auth';
const USER_COOKIE_KEY = 'profile';

function getCookieSecret(): string {
	if (typeof process !== 'undefined' && process.env?.COOKIE_SECRET) {
		return process.env.COOKIE_SECRET;
	}
	return 'default-secret-change-in-production';
}

const cookieSecret = getCookieSecret();

export interface SetAuthCookieOptions {
	trim?: boolean;
	maxAge?: number;
	secure?: boolean;
}

export function setAuthCookie(cookies: Cookies, token: string, options?: SetAuthCookieOptions) {
	const {
		trim = false,
		maxAge = 60 * 60 * 24 * 365,
		secure = import.meta.env.PROD
	} = options ?? {};

	const value = trim ? token.trim() : token;
	cookies.set(AUTH_KEY, value, {
		path: '/',
		httpOnly: true,
		secure,
		sameSite: 'lax',
		maxAge
	});
}

export function deleteAuthCookie(cookies: Cookies) {
	cookies.delete(AUTH_KEY, { path: '/' });
}

export function getAuthCookie(cookies: Cookies) {
	return cookies.get(AUTH_KEY);
}

export { AUTH_KEY };

function sign(data: string): string {
	return createHmac('sha256', cookieSecret).update(data).digest('hex');
}

function encode(userData: User): string {
	const json = JSON.stringify(userData);
	const signature = sign(json);
	return `${json}.${signature}`;
}

function decode(cookieValue: string): User | null {
	const lastDotIndex = cookieValue.lastIndexOf('.');
	if (lastDotIndex === -1) {
		return null;
	}
	const json = cookieValue.slice(0, lastDotIndex);
	const signature = cookieValue.slice(lastDotIndex + 1);
	const expectedSignature = sign(json);
	if (!timingSafeEqual(Buffer.from(signature), Buffer.from(expectedSignature))) {
		return null;
	}
	try {
		return JSON.parse(json) as User;
	} catch {
		return null;
	}
}

export function getUserCookie(cookies: Cookies): User | undefined {
	const value = cookies.get(USER_COOKIE_KEY);
	if (!value) {
		return undefined;
	}
	return decode(value) ?? undefined;
}

export function setUserCookie(cookies: Cookies, userData: User) {
	const value = encode(userData);
	cookies.set(USER_COOKIE_KEY, value, {
		path: '/',
		httpOnly: true,
		secure: import.meta.env.PROD,
		sameSite: 'lax',
		maxAge: 60 * 60 * 24 * 365
	});
}

export function deleteUserCookie(cookies: Cookies) {
	cookies.delete(USER_COOKIE_KEY, { path: '/' });
}
