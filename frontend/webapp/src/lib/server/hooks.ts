import { redirect } from '@sveltejs/kit';
import type { UserProfile } from '@savetoink/shared';
import { validateEnv } from '$lib/server/validateEnv';
import {
	getAuthCookie,
	getRefreshCookie,
	setAuthCookie,
	setRefreshCookie,
	getUserCookie,
	setUserCookie,
	clearAuthCookies,
	setRedirectToCookie
} from '$lib/server/cookies';
import { isAuthenticatedPath } from '$lib/consts';
import { getProfile, refreshToken } from '$lib/server/apiClient';
import type { Handle } from '@sveltejs/kit';

/**
 * Attempt to load user from the signed profile cookie cache.
 */
async function loadCachedUser(
	cookies: Parameters<Handle>['0']['event']['cookies']
): Promise<UserProfile | undefined> {
	return await getUserCookie(cookies);
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
		return true;
	} catch {
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
		return undefined;
	}

	try {
		const refreshed = await refreshToken(event, refresh);
		setAuthCookie(event.cookies, refreshed.access_token, {
			maxAge: refreshed.expires_in
		});
		if (refreshed.refresh_token) {
			setRefreshCookie(event.cookies, refreshed.refresh_token);
		}
		return refreshed.access_token;
	} catch {
		return undefined;
	}
}

export const handle: Handle = async ({ event, resolve }) => {
	validateEnv();

	event.locals.auth = getAuthCookie(event.cookies);
	event.locals.isLoggedIn = false;
	event.locals.user = undefined;

	// --- No auth cookie: try refresh before giving up ---
	if (!event.locals.auth) {
		const newToken = await attemptTokenRefresh(event);
		if (newToken) {
			event.locals.auth = newToken;
			await fetchProfile(event);
		}
		return resolveOrRedirect(event, resolve);
	}

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

	// Step 3: Token expired — try refresh
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
	clearAuthCookies(event.cookies);
	return resolveOrRedirect(event, resolve);
};

function resolveOrRedirect(
	event: Parameters<Handle>['0']['event'],
	resolve: Parameters<Handle>['0']['resolve']
) {
	if (isAuthenticatedPath(event.url.pathname) && !event.locals.isLoggedIn) {
		setRedirectToCookie(event.cookies, event.url.pathname + event.url.search);
		return redirect(303, '/account');
	}

	return resolve(event);
}
