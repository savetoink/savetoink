import { sequence } from '@sveltejs/kit/hooks';
import { handleErrorWithSentry, sentryHandle } from '@sentry/sveltekit';
import { redirect } from '@sveltejs/kit';
import { validateEnv } from '$lib/server/validateEnv';
import { getJwtCookie, deleteJwtCookie } from '$lib/server/cookies';
import { GET } from '$lib/server/apiClient';
import { isAuthenticatedPath } from '$lib/consts';

import type { Handle } from '@sveltejs/kit';

export const handle: Handle = sequence(sentryHandle(), async ({ event, resolve }) => {
	validateEnv();

	event.locals.jwt = getJwtCookie(event.cookies);
	event.locals.isLoggedIn = false;
	event.locals.user = undefined;

	if (event.locals.jwt) {
		try {
			const profile = await GET(event.fetch, '/v1/user/profile', event.locals.jwt);
			event.locals.user = {
				account: profile.account,
				email: profile.email,
				deviceEmail: profile.device_email,
				autoSend: profile.auto_send
			};
			event.locals.isLoggedIn = true;
		} catch {
			deleteJwtCookie(event.cookies);
			event.locals.jwt = undefined;
			if (isAuthenticatedPath(event.url.pathname)) {
				return redirect(303, '/account');
			}
		}
	}

	if (isAuthenticatedPath(event.url.pathname) && !event.locals.isLoggedIn) {
		return redirect(303, '/account');
	}

	return resolve(event);
});

export const handleError = handleErrorWithSentry();
