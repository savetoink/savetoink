import { error, redirect } from '@sveltejs/kit';
import { POST } from '$lib/server/apiClient';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ fetch, url, cookies, locals }) => {
	const code = url.searchParams.get('code');

	if (!code) {
		error(500, 'missing code');
	}

	const response = await POST(
		fetch,
		'/v1/auth/token',
		{
			code: code,
			redirect_uri: `${url.origin}/auth/callback`
		},
		locals.jwt // remove and check
	);

	if (!response.access_token) {
		console.error(response);
		error(500, 'Auth0 exchange_failed');
	}

 // create a set cookie shared util
	// Store the JWT in an HTTP-only cookie
	cookies.set('jwt', response.access_token, {
		path: '/',
		httpOnly: true, // JavaScript can't access it
		secure: true, // Only sent over HTTPS in production
		sameSite: 'lax',
		maxAge: 60 * 60 * 24 // 24 hours
	});

	redirect(303, '/');
};
