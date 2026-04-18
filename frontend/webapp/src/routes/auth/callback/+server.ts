import { error, redirect } from '@sveltejs/kit';
import { env as publicEnv } from '$env/dynamic/public';
import { exchangeCodeForToken, getProfile } from '$lib/server/apiClient';
import { setTokenCookies, setUserCookie, getValidRedirectUrl } from '$lib/server/cookies';
import type { RequestHandler, RequestEvent } from './$types';

export const GET: RequestHandler = async ({ fetch, url, cookies, request, getClientAddress }) => {
	const code = url.searchParams.get('code');

	if (!code) {
		error(500, 'missing code');
	}

	const response = await exchangeCodeForToken(
		{ fetch, request, getClientAddress } as RequestEvent,
		code,
		`${publicEnv.PUBLIC_APP_URL || url.origin}/auth/callback`
	);

	if (!response.access_token) {
		error(500, 'Auth0 exchange_failed');
	}

	setTokenCookies(cookies, response);

	const profile = await getProfile({
		locals: { auth: response.access_token },
		fetch,
		request,
		getClientAddress
	} as RequestEvent);
	await setUserCookie(cookies, {
		account: profile.account,
		email: profile.email,
		device_email: profile.device_email,
		auto_send: profile.auto_send
	});

	redirect(303, getValidRedirectUrl(cookies) || '/');
};
