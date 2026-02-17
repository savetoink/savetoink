import type { Handle } from '@sveltejs/kit';
import { error } from '@sveltejs/kit';
import { PUBLIC_API_URL } from '$env/static/public';

export const handle: Handle = async ({ event, resolve }) => {
	if (!PUBLIC_API_URL) {
		error(500, 'PUBLIC_API_URL is not defined');
	}

	const jwt = event.cookies.get('jwt');
	event.locals.jwt = jwt;

	return resolve(event);
};
