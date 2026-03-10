import { error, redirect } from '@sveltejs/kit';
import { exchangeCodeForToken, getProfile } from '$lib/server/apiClient';
import { setAuthCookie, setUserCookie } from '$lib/server/cookies';
import type { RequestHandler, RequestEvent } from './$types';

export const GET: RequestHandler = async ({ fetch, url, cookies, request }) => {
	const code = url.searchParams.get('code');

	if (!code) {
		error(500, 'missing code');
	}

	const response = await exchangeCodeForToken(
		{ fetch, request } as RequestEvent,
		code,
		`${url.origin}/auth/callback`
	);

	if (!response.access_token) {
		error(500, 'Auth0 exchange_failed');
	}

	setAuthCookie(cookies, response.access_token, {
		maxAge: response.expires_in
	});

	const profile = await getProfile({
		locals: { auth: response.access_token },
		fetch,
		request
	} as RequestEvent);
	await setUserCookie(cookies, {
		account: profile.account,
		email: profile.email,
		device_email: profile.device_email,
		auto_send: profile.auto_send
	});

	redirect(303, '/');
};
