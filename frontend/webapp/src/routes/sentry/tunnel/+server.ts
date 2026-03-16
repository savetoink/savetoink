import { env as publicEnv } from '$env/dynamic/public';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ request }) => {
	const sentryDsnUrl = new URL(publicEnv.PUBLIC_SENTRY_DSN);
	const sentryIngestUrl = `${sentryDsnUrl.protocol}//${sentryDsnUrl.host}/api${sentryDsnUrl.pathname}`;

	const body = await request.text();
	const headers = new Headers();

	for (const [key, value] of request.headers.entries()) {
		if (key.toLowerCase() !== 'host') {
			headers.set(key, value);
		}
	}

	headers.set('Host', sentryDsnUrl.host);

	const response = await fetch(`${sentryIngestUrl}/envelope/`, {
		method: 'POST',
		headers,
		body
	});

	return new Response(response.body, {
		status: response.status,
		headers: response.headers
	});
};
