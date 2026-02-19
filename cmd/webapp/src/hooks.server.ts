import { redirect } from '@sveltejs/kit';
import { validateEnv } from '$lib/server/validateEnv';
import { getJwtCookie } from '$lib/server/cookies';
import { isAuthenticatedPath } from '$lib/consts';

import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
	validateEnv();

	event.locals.jwt = getJwtCookie(event.cookies);

	if (isAuthenticatedPath(event.url.pathname) && !event.locals?.jwt) {
		return redirect(303, '/login');
	}

	return resolve(event);
};
