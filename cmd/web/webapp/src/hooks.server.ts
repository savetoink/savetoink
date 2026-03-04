import { sequence } from '@sveltejs/kit/hooks';
import { handleErrorWithSentry, sentryHandle } from '@sentry/sveltekit';
import { redirect } from '@sveltejs/kit';
import { validateEnv } from '$lib/server/validateEnv';
import { getAuthCookie, getUserCookie, setUserCookie } from '$lib/server/cookies';
import { isAuthenticatedPath } from '$lib/consts';
import { getProfile } from '$lib/server/apiClient';

import type { Handle } from '@sveltejs/kit';

export const handle: Handle = sequence(sentryHandle(), async ({ event, resolve }) => {
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
				deviceEmail: userData.deviceEmail,
				autoSend: userData.autoSend
			};
			event.locals.isLoggedIn = true;
		} else {
			// get profile data from API if user is authenticated but no cookie is found
			try {
				const profileData = await getProfile(event.fetch, event.locals.auth);
				event.locals.user = {
					account: profileData.account,
					email: profileData.email,
					deviceEmail: profileData.deviceEmail,
					autoSend: profileData.autoSend
				};
				event.locals.isLoggedIn = true;
				await setUserCookie(event.cookies, profileData);
			} catch {
				// Profile fetch failed, stay logged out
			}
		}
	}

	if (isAuthenticatedPath(event.url.pathname) && !event.locals.isLoggedIn) {
		return redirect(303, '/account');
	}

	return resolve(event);
});

export const handleError = handleErrorWithSentry();
