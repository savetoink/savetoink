import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
	const jwt = event.cookies.get('jwt');
	event.locals.jwt = jwt;

	return resolve(event);
};
