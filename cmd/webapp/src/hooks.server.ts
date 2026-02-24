import { sequence } from '@sveltejs/kit/hooks';
import { handleErrorWithSentry, sentryHandle } from '@sentry/sveltekit';
import { redirect } from '@sveltejs/kit';
import { validateEnv } from '$lib/server/validateEnv';
import { getJwtCookie } from '$lib/server/cookies';
import { isAuthenticatedPath } from '$lib/consts';

import type { Handle } from '@sveltejs/kit';

export const handle: Handle = sequence(sentryHandle(), async ({ event, resolve }) => {
	validateEnv();

	event.locals.jwt = getJwtCookie(event.cookies);
	event.locals.isLoggedIn = !!event.locals.jwt;

	if (isAuthenticatedPath(event.url.pathname) && !event.locals?.jwt) {
		return redirect(303, '/account');
	}

	return resolve(event);
});

export const handleError = handleErrorWithSentry();
