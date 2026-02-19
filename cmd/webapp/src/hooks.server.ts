import type { Handle } from '@sveltejs/kit';
import { validateEnv } from '$lib/server/validateEnv';
import { getJwtCookie } from '$lib/server/cookies';

export const handle: Handle = async ({ event, resolve }) => {
	validateEnv();

	event.locals.jwt = getJwtCookie(event.cookies);

	return resolve(event);
};
