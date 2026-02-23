import { PUBLIC_SENTRY_DSN } from '$env/static/public';
import type { RequestHandler } from './$types';

const sentryDsnUrl = new URL(PUBLIC_SENTRY_DSN);
const sentryIngestUrl = `${sentryDsnUrl.protocol}//${sentryDsnUrl.host}${sentryDsnUrl.pathname}`;

export const POST: RequestHandler = async ({ request }) => {
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
