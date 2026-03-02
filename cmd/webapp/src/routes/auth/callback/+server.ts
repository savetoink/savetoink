import { error, redirect } from '@sveltejs/kit';
import { POST, GET as apiGet } from '$lib/server/apiClient';
import { setAuthCookie, setUserCookie } from '$lib/server/cookies';
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

	setAuthCookie(cookies, response.access_token, {
		maxAge: response?.expires_in
	});

	const profile = await apiGet(fetch, '/v1/user/profile', response.access_token);
	await setUserCookie(cookies, {
		account: profile.account,
		email: profile.email,
		deviceEmail: profile.device_email,
		autoSend: profile.auto_send
	});

	redirect(303, '/');
};
