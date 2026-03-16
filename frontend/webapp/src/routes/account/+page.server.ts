import { fail, redirect } from '@sveltejs/kit';
import {
	PUBLIC_AUTH0_CLIENT_ID,
	PUBLIC_AUTH0_DOMAIN,
	PUBLIC_AUTH_BACKEND,
	PUBLIC_APP_URL
} from '$env/static/public';
import {
	AUTH_KEY,
	setAuthCookie,
	deleteAuthCookie,
	setUserCookie,
	deleteUserCookie
} from '$lib/server/cookies';
import { getProfile, updateDevice, deleteDevice } from '$lib/server/apiClient';
import { ApiError, DeviceDomains, Auth0 } from '@savetoink/shared';
import type { UserProfile } from '@savetoink/shared';
import type { Actions, PageServerLoad, RequestEvent } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	return { user: locals.user };
};

export const actions: Actions = {
	clean: async ({ cookies }) => {
		deleteAuthCookie(cookies);
		deleteUserCookie(cookies);

		if (PUBLIC_AUTH_BACKEND === Auth0) {
			const auth0LogoutUrl = new URL(`https://${PUBLIC_AUTH0_DOMAIN}/logout`);
			auth0LogoutUrl.searchParams.set('client_id', PUBLIC_AUTH0_CLIENT_ID);
			auth0LogoutUrl.searchParams.set('returnTo', `${PUBLIC_APP_URL}/account`);
			redirect(303, auth0LogoutUrl.toString());
		}

		redirect(303, '/account');
	},
	save: async ({ cookies, request, fetch, getClientAddress }) => {
		const data = await request.formData();
		const token = data.get(AUTH_KEY);

		if (!token || typeof token !== 'string' || token.trim() === '') {
			return fail(400, { error: 'api key is required' });
		}

		let profile: UserProfile;
		try {
			profile = await getProfile({
				locals: { auth: token },
				fetch,
				request,
				getClientAddress
			} as RequestEvent);
		} catch (err) {
			if (err instanceof ApiError) {
				return fail(400, { error: 'Unauthorized: ' + err.message });
			}
			return fail(500, { error: 'Failed to validate api key: ' + err });
		}

		setAuthCookie(cookies, token, { trim: true });
		await setUserCookie(cookies, {
			account: profile.account,
			email: profile.email,
			device_email: profile.device_email,
			auto_send: profile.auto_send
		});

		redirect(303, '/');
	},
	updateAutoSend: async ({ locals, request, fetch, cookies, getClientAddress }) => {
		const data = await request.formData();
		const autoSend = data.get('autoSend');

		await updateDevice(
			{ locals, fetch, request, getClientAddress } as RequestEvent,
			locals.user?.device_email || '',
			autoSend === 'on'
		);

		const updatedProfile = await getProfile({
			locals,
			fetch,
			request,
			getClientAddress
		} as RequestEvent);
		await setUserCookie(cookies, {
			account: updatedProfile.account,
			email: updatedProfile.email,
			device_email: updatedProfile.device_email,
			auto_send: updatedProfile.auto_send
		});
		return { success: true };
	},
	updateProfile: async ({ locals, request, fetch, cookies, getClientAddress }) => {
		const data = await request.formData();
		const deviceEmail = data.get('deviceEmail');
		const autoSend = data.get('autoSend');

		if (typeof deviceEmail === 'string' && deviceEmail.trim() !== '') {
			const parts = deviceEmail.split('@');
			if (parts.length !== 2 || !parts[1]) {
				return fail(400, { error: 'invalid email format' });
			}

			const domain = '@' + parts[1];
			if (!DeviceDomains.includes(domain)) {
				return fail(400, { error: 'kindle email domain must be ' + DeviceDomains.join(' or ') });
			}
		}

		await updateDevice(
			{ locals, fetch, request, getClientAddress } as RequestEvent,
			deviceEmail ? deviceEmail.toString() : '',
			autoSend === 'on'
		);

		const updatedProfile = await getProfile({
			locals,
			fetch,
			request,
			getClientAddress
		} as RequestEvent);
		await setUserCookie(cookies, {
			account: updatedProfile.account,
			email: updatedProfile.email,
			device_email: updatedProfile.device_email,
			auto_send: updatedProfile.auto_send
		});
		redirect(303, '/account');
	},
	deleteDevice: async ({ locals, request, fetch, cookies, getClientAddress }) => {
		await deleteDevice({ locals, fetch, request, getClientAddress } as RequestEvent);
		const profile = await getProfile({ locals, fetch, request, getClientAddress } as RequestEvent);
		await setUserCookie(cookies, {
			account: profile.account,
			email: profile.email,
			device_email: profile.device_email,
			auto_send: profile.auto_send
		});
		redirect(303, '/account');
	}
} satisfies Actions;
