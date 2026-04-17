import { env as privateEnv } from '$env/dynamic/private';
import type { Cookies } from '@sveltejs/kit';
import type { UserProfile } from '@savetoink/shared';

const AUTH_KEY = 'auth';
const REFRESH_KEY = 'refresh';
const USER_COOKIE_KEY = 'profile';
const REDIRECT_TO_KEY = 'redirect_to';

function getCookieSecret(): string {
	if (privateEnv.COOKIE_SECRET) {
		return privateEnv.COOKIE_SECRET;
	}
	return 'default-secret-change-in-production';
}

function getSecure(): boolean {
	return !!privateEnv.COOKIE_SECRET;
}

const cookieSecret = getCookieSecret();

async function getCryptoKey(): Promise<CryptoKey> {
	const encoder = new TextEncoder();
	const keyData = encoder.encode(cookieSecret);
	return await crypto.subtle.importKey('raw', keyData, { name: 'HMAC', hash: 'SHA-256' }, false, [
		'sign',
		'verify'
	]);
}

function timingSafeEqual(a: Uint8Array, b: Uint8Array): boolean {
	if (a.length !== b.length) {
		return false;
	}
	let result = 0;
	for (let i = 0; i < a.length; i++) {
		result |= a[i] ^ b[i];
	}
	return result === 0;
}

export interface SetAuthCookieOptions {
	trim?: boolean;
	maxAge?: number;
	secure?: boolean;
}

export function setAuthCookie(cookies: Cookies, token: string, options?: SetAuthCookieOptions) {
	const { trim = false, maxAge = 60 * 60 * 24 * 365, secure = getSecure() } = options ?? {};

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
	cookies.delete(AUTH_KEY, { path: '/', secure: getSecure() });
}

export function getAuthCookie(cookies: Cookies) {
	return cookies.get(AUTH_KEY);
}

export { AUTH_KEY };

export function setRefreshCookie(cookies: Cookies, token: string) {
	cookies.set(REFRESH_KEY, token, {
		path: '/',
		httpOnly: true,
		secure: getSecure(),
		sameSite: 'lax',
		maxAge: 60 * 60 * 24 * 7 // 7 days (matches PASETO refresh TTL)
	});
}

export function deleteRefreshCookie(cookies: Cookies) {
	cookies.delete(REFRESH_KEY, { path: '/', secure: getSecure() });
}

export function getRefreshCookie(cookies: Cookies) {
	return cookies.get(REFRESH_KEY);
}

export { REFRESH_KEY };

async function sign(data: string): Promise<string> {
	const key = await getCryptoKey();
	const encoder = new TextEncoder();
	const dataBuffer = encoder.encode(data);
	const signature = await crypto.subtle.sign('HMAC', key, dataBuffer);
	const signatureArray = Array.from(new Uint8Array(signature));
	return signatureArray.map((b) => b.toString(16).padStart(2, '0')).join('');
}

async function encode(userData: UserProfile): Promise<string> {
	const json = JSON.stringify(userData);
	const signature = await sign(json);
	return `${json}.${signature}`;
}

async function decode(cookieValue: string): Promise<UserProfile | null> {
	const lastDotIndex = cookieValue.lastIndexOf('.');
	if (lastDotIndex === -1) {
		return null;
	}
	const json = cookieValue.slice(0, lastDotIndex);
	const signature = cookieValue.slice(lastDotIndex + 1);
	const expectedSignature = await sign(json);

	const sigBuffer = new Uint8Array(signature.length / 2);
	for (let i = 0; i < signature.length; i += 2) {
		sigBuffer[i / 2] = parseInt(signature.substring(i, i + 2), 16);
	}

	const expectedSigBuffer = new Uint8Array(expectedSignature.length / 2);
	for (let i = 0; i < expectedSignature.length; i += 2) {
		expectedSigBuffer[i / 2] = parseInt(expectedSignature.substring(i, i + 2), 16);
	}

	if (!timingSafeEqual(sigBuffer, expectedSigBuffer)) {
		return null;
	}
	try {
		return JSON.parse(json) as UserProfile;
	} catch {
		return null;
	}
}

export async function getUserCookie(cookies: Cookies): Promise<UserProfile | undefined> {
	const value = cookies.get(USER_COOKIE_KEY);
	if (!value) {
		return undefined;
	}
	return (await decode(value)) ?? undefined;
}

export async function setUserCookie(cookies: Cookies, userData: UserProfile) {
	const value = await encode(userData);
	cookies.set(USER_COOKIE_KEY, value, {
		path: '/',
		httpOnly: true,
		secure: getSecure(),
		sameSite: 'lax',
		maxAge: 60 * 60 * 24 * 365
	});
}

export function deleteUserCookie(cookies: Cookies) {
	cookies.delete(USER_COOKIE_KEY, { path: '/', secure: getSecure() });
}

export function setRedirectToCookie(cookies: Cookies, url: string) {
	cookies.set(REDIRECT_TO_KEY, url, {
		path: '/',
		httpOnly: true,
		secure: getSecure(),
		sameSite: 'lax',
		maxAge: 300 // 5 minutes
	});
}

function getRedirectToCookie(cookies: Cookies): string | undefined {
	const value = cookies.get(REDIRECT_TO_KEY);
	if (!value) {
		return undefined;
	}
	return value;
}

function deleteRedirectToCookie(cookies: Cookies) {
	cookies.delete(REDIRECT_TO_KEY, { path: '/', secure: getSecure() });
}

export { REDIRECT_TO_KEY };

/**
 * Validates a redirect URL to prevent open redirect vulnerabilities.
 * Only allows relative URLs starting with '/' to prevent external redirects.
 */
function isValidRedirectUrl(url: string | undefined): boolean {
	if (!url) {
		return false;
	}

	if (!url.startsWith('/')) {
		return false;
	}

	if (url.startsWith('//')) {
		return false;
	}

	try {
		// Try to parse as URL - this should fail for relative paths
		new URL(url);
		return false;
	} catch {
		// Expected for relative URLs - valid
	}

	return true;
}

/**
 * Gets the redirect URL from cookie, validates it, and deletes the cookie.
 * Returns the valid redirect URL or undefined if no valid redirect exists.
 * Deletes the cookie in both valid and invalid cases for security.
 */
export function getValidRedirectUrl(cookies: Cookies): string | undefined {
	const redirectTo = getRedirectToCookie(cookies);
	if (!redirectTo) {
		return undefined;
	}

	if (isValidRedirectUrl(redirectTo)) {
		deleteRedirectToCookie(cookies);
		return redirectTo;
	}

	// Invalid URL - delete the cookie for security
	deleteRedirectToCookie(cookies);
	return undefined;
}
