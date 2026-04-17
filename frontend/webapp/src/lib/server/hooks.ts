import { redirect } from '@sveltejs/kit';
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
	setRedirectToCookie
} from '$lib/server/cookies';
import { isAuthenticatedPath } from '$lib/consts';
import { getProfile, refreshToken } from '$lib/server/apiClient';
import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
	validateEnv();

	event.locals.auth = getAuthCookie(event.cookies);
	event.locals.isLoggedIn = false;
	event.locals.user = undefined;

	if (event.locals.auth) {
		const userData = await getUserCookie(event.cookies);
		if (userData) {
			event.locals.user = {
				account: userData.account,
				email: userData.email,
				device_email: userData.device_email,
				auto_send: userData.auto_send
			};
			event.locals.isLoggedIn = true;
		} else {
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
			} catch {
				const refresh = getRefreshCookie(event.cookies);
				if (refresh) {
					try {
						const refreshed = await refreshToken(event, refresh);
						setAuthCookie(event.cookies, refreshed.access_token, {
							maxAge: refreshed.expires_in
						});
						if (refreshed.refresh_token) {
							setRefreshCookie(event.cookies, refreshed.refresh_token);
						}
						event.locals.auth = refreshed.access_token;

						const profileData = await getProfile(event);
						event.locals.user = {
							account: profileData.account,
							email: profileData.email,
							device_email: profileData.device_email,
							auto_send: profileData.auto_send
						};
						event.locals.isLoggedIn = true;
						await setUserCookie(event.cookies, profileData);
					} catch {
						deleteAuthCookie(event.cookies);
						deleteRefreshCookie(event.cookies);
					}
				}
			}
		}
	}

	if (isAuthenticatedPath(event.url.pathname) && !event.locals.isLoggedIn) {
		setRedirectToCookie(event.cookies, event.url.pathname + event.url.search);
		return redirect(303, '/account');
	}

	return resolve(event);
};
