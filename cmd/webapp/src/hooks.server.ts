import type { Handle } from '@sveltejs/kit';
import { error } from '@sveltejs/kit';
import {
	PUBLIC_API_URL,
	PUBLIC_AUTH_BACKEND,
	PUBLIC_AUTH0_CLIENT_ID,
	PUBLIC_AUTH0_DOMAIN,
	PUBLIC_AUTH0_AUDIENCE
} from '$env/static/public';
import { getJwtCookie } from '$lib/server/cookies';

export const handle: Handle = async ({ event, resolve }) => {
	const requiredEnvVars = [
		{ name: 'PUBLIC_API_URL', value: PUBLIC_API_URL },
		{ name: 'PUBLIC_AUTH_BACKEND', value: PUBLIC_AUTH_BACKEND }
	];

	if (PUBLIC_AUTH_BACKEND === 'auth0') {
		requiredEnvVars.push(
			{ name: 'PUBLIC_AUTH0_CLIENT_ID', value: PUBLIC_AUTH0_CLIENT_ID },
			{ name: 'PUBLIC_AUTH0_DOMAIN', value: PUBLIC_AUTH0_DOMAIN },
			{ name: 'PUBLIC_AUTH0_AUDIENCE', value: PUBLIC_AUTH0_AUDIENCE }
		);
	}

	const missing = requiredEnvVars.filter(({ value }) => !value).map(({ name }) => name);

	if (missing.length > 0) {
		error(500, `Required environment variables are not defined: ${missing.join(', ')}`);
	}

	event.locals.jwt = getJwtCookie(event.cookies);

	return resolve(event);
};
