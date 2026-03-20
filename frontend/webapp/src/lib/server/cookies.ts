import { env as privateEnv } from '$env/dynamic/private';
import type { Cookies } from '@sveltejs/kit';
import type { UserProfile } from '@savetoink/shared';

const AUTH_KEY = 'auth';
const USER_COOKIE_KEY = 'profile';
const REFERRER_COOKIE_KEY = 'referrer';

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

export function setReferrerCookie(cookies: Cookies, referrer: string) {
	cookies.set(REFERRER_COOKIE_KEY, referrer, {
		path: '/',
		httpOnly: false,
		secure: getSecure(),
		sameSite: 'lax',
		maxAge: 60 * 10
	});
}

export function getReferrerCookie(cookies: Cookies): string | undefined {
	return cookies.get(REFERRER_COOKIE_KEY);
}

export function deleteReferrerCookie(cookies: Cookies) {
	cookies.delete(REFERRER_COOKIE_KEY, { path: '/', secure: getSecure() });
}
