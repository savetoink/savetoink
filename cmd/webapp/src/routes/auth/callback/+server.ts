import { error, redirect } from '@sveltejs/kit';
import { POST } from '$lib/server/apiClient';
import { setJwtCookie } from '$lib/server/cookies';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ fetch, url, cookies }) => {
	const code = url.searchParams.get('code');

	if (!code) {
		error(500, 'missing code');
	}

	const response = await POST(fetch, '/v1/auth/token', {
		code: code,
		redirect_uri: `${url.origin}/auth/callback`
	});

	if (!response.access_token) {
		console.error(response);
		error(500, 'Auth0 exchange_failed');
	}

	setJwtCookie(cookies, response.access_token, {
		secure: true,
		maxAge: 60 * 60 * 24
	});

	redirect(303, '/');
};
