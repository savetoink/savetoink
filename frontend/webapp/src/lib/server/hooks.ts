import { redirect } from '@sveltejs/kit';
import { validateEnv } from '$lib/server/validateEnv';
import {
	getAuthCookie,
	getUserCookie,
	setUserCookie,
	setRedirectToCookie
} from '$lib/server/cookies';
import { isAuthenticatedPath } from '$lib/consts';
import { getProfile } from '$lib/server/apiClient';
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
			// get profile data from API if user is authenticated but no cookie is found
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
				// Profile fetch failed, stay logged out
			}
		}
	}

	if (isAuthenticatedPath(event.url.pathname) && !event.locals.isLoggedIn) {
		setRedirectToCookie(event.cookies, event.url.pathname + event.url.search);
		return redirect(303, '/account');
	}

	return resolve(event);
};
