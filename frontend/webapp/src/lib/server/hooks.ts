import { redirect } from '@sveltejs/kit';
import type { UserProfile } from '@savetoink/shared';
import { validateEnv } from '$lib/server/validateEnv';
import {
	getAuthCookie,
	getRefreshCookie,
	setAuthCookie,
	setRefreshCookie,
	deleteAuthCookie,
	deleteRefreshCookie,
	getUserCookie,
	setUserCookie,
	deleteUserCookie,
	setRedirectToCookie
} from '$lib/server/cookies';
import { isAuthenticatedPath } from '$lib/consts';
import { getProfile, refreshToken } from '$lib/server/apiClient';
import { ApiError } from '@savetoink/shared';
import type { Handle } from '@sveltejs/kit';

const DEBUG = true;

function log(tag: string, ...args: unknown[]) {
	if (!DEBUG) return;
	console.log(`[auth:${tag}]`, ...args);
}

function logError(tag: string, ...args: unknown[]) {
	console.error(`[auth:${tag}]`, ...args);
}

/**
 * Attempt to load user from the signed profile cookie cache.
 * Returns the user profile on success, or undefined.
 */
async function loadCachedUser(
	cookies: Parameters<Handle>['0']['event']['cookies']
): Promise<UserProfile | undefined> {
	const userData = await getUserCookie(cookies);
	if (!userData) {
		log('cache', 'no user cookie found');
		return undefined;
	}
	log('cache', 'user cookie found', { account: userData.account });
	return userData;
}

/**
 * Fetch the user profile from the API and populate locals + cookie cache.
 * Returns true on success, false on failure.
 */
async function fetchProfile(event: Parameters<Handle>['0']['event']): Promise<boolean> {
	try {
		const profileData = await getProfile(event);
		event.locals.user = {
			account: profileData.account,
			email: profileData.email,
			device_email: profileData.device_email,
			auto_send: profileData.auto_send
		};
		event.locals.isLoggedIn = true;
		await setUserCookie(event.cookies, profileData);
		log('profile', 'fetched successfully', { account: profileData.account });
		return true;
	} catch (e) {
		const msg = e instanceof Error ? e.message : String(e);
		const status = e instanceof ApiError ? e.status : undefined;
		logError('profile', 'fetch failed', { status, msg });
		return false;
	}
}

/**
 * Attempt to refresh the access token using the refresh cookie.
 * Returns the new access token on success, or undefined on failure.
 */
async function attemptTokenRefresh(
	event: Parameters<Handle>['0']['event']
): Promise<string | undefined> {
	const refresh = getRefreshCookie(event.cookies);
	if (!refresh) {
		log('refresh', 'no refresh cookie available');
		return undefined;
	}

	try {
		log('refresh', 'attempting token refresh');
		const refreshed = await refreshToken(event, refresh);
		log('refresh', 'token refreshed successfully', {
			hasNewRefreshToken: !!refreshed.refresh_token,
			expiresIn: refreshed.expires_in
		});

		setAuthCookie(event.cookies, refreshed.access_token, {
			maxAge: refreshed.expires_in
		});
		if (refreshed.refresh_token) {
			setRefreshCookie(event.cookies, refreshed.refresh_token);
		}

		return refreshed.access_token;
	} catch (e) {
		const msg = e instanceof Error ? e.message : String(e);
		const status = e instanceof ApiError ? e.status : undefined;
		logError('refresh', 'token refresh failed', { status, msg });
		return undefined;
	}
}

/**
 * Clear all authentication cookies (full logout).
 */
function clearAuthCookies(cookies: Parameters<Handle>['0']['event']['cookies']) {
	log('cookies', 'clearing all auth cookies');
	deleteAuthCookie(cookies);
	deleteRefreshCookie(cookies);
	deleteUserCookie(cookies);
}

export const handle: Handle = async ({ event, resolve }) => {
	validateEnv();

	event.locals.auth = getAuthCookie(event.cookies);
	event.locals.isLoggedIn = false;
	event.locals.user = undefined;

	// --- No auth cookie: try refresh before giving up ---
	if (!event.locals.auth) {
		log('init', 'no auth cookie', { path: event.url.pathname });

		const newToken = await attemptTokenRefresh(event);
		if (!newToken) {
			return resolveOrRedirect(event, resolve);
		}

		log('init', 'restored auth via refresh token');
		event.locals.auth = newToken;

		if (await fetchProfile(event)) {
			return resolveOrRedirect(event, resolve);
		}

		logError('init', 'profile fetch failed after restoring token, clearing cookies');
		clearAuthCookies(event.cookies);
		return resolveOrRedirect(event, resolve);
	}

	log('init', 'auth cookie present', { path: event.url.pathname });

	// Step 1: Try user cookie cache
	const cachedUser = await loadCachedUser(event.cookies);
	if (cachedUser) {
		event.locals.user = cachedUser;
		event.locals.isLoggedIn = true;
		return resolveOrRedirect(event, resolve);
	}

	// Step 2: Try fetching profile with current token
	if (await fetchProfile(event)) {
		return resolveOrRedirect(event, resolve);
	}

	// Step 3: Token might be expired — try refresh
	const newToken = await attemptTokenRefresh(event);
	if (!newToken) {
		clearAuthCookies(event.cookies);
		return resolveOrRedirect(event, resolve);
	}

	// Step 4: Retry profile fetch with new token
	event.locals.auth = newToken;
	if (await fetchProfile(event)) {
		return resolveOrRedirect(event, resolve);
	}

	// Profile fetch failed even with a fresh token — clear everything
	logError('init', 'profile fetch failed even after refresh, clearing cookies');
	clearAuthCookies(event.cookies);
	return resolveOrRedirect(event, resolve);
};

/**
 * Resolve the request, redirecting to /account if the path requires auth
 * and the user is not logged in.
 */
function resolveOrRedirect(
	event: Parameters<Handle>['0']['event'],
	resolve: Parameters<Handle>['0']['resolve']
) {
	if (isAuthenticatedPath(event.url.pathname) && !event.locals.isLoggedIn) {
		log('redirect', 'unauthenticated access to protected path', { path: event.url.pathname });
		setRedirectToCookie(event.cookies, event.url.pathname + event.url.search);
		return redirect(303, '/account');
	}

	return resolve(event);
}
