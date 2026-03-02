import { error, fail, redirect } from '@sveltejs/kit';
import {
	AUTH_KEY,
	setAuthCookie,
	deleteAuthCookie,
	setUserCookie,
	deleteUserCookie
} from '$lib/server/cookies';
import { GET, PUT, DELETE } from '$lib/server/apiClient';
import { DeviceDomains } from '$lib/consts';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	return { user: locals.user };
};

export const actions: Actions = {
	save: async ({ cookies, request, fetch }) => {
		const data = await request.formData();
		const token = data.get(AUTH_KEY);

		if (!token || typeof token !== 'string' || token.trim() === '') {
			return fail(400, { error: 'api key is required' });
		}

		let profile;
		try {
			profile = await GET(fetch, '/v1/user/profile', token);
		} catch (err) {
			console.error('failed to validate api key', err);
			return fail(400, { error: 'Invalid API key' });
		}

		setAuthCookie(cookies, token, { trim: true });
		await setUserCookie(cookies, {
			account: profile.account,
			email: profile.email,
			deviceEmail: profile.device_email,
			autoSend: profile.auto_send
		});

		redirect(303, '/');
	},
	clean: async ({ cookies }) => {
		deleteAuthCookie(cookies);
		deleteUserCookie(cookies);
		redirect(303, '/account');
	},
	updateProfile: async ({ locals, request, fetch, cookies }) => {
		const data = await request.formData();
		const deviceEmail = data.get('deviceEmail');
		const autoSend = data.get('autoSend');

		if (deviceEmail && typeof deviceEmail === 'string' && deviceEmail.trim() !== '') {
			const parts = deviceEmail.split('@');
			if (parts.length !== 2 || !parts[1]) {
				return error(400, 'invalid email format');
			}

			const domain = '@' + parts[1];
			if (!DeviceDomains.includes(domain)) {
				return error(400, 'kindle email domain must be ' + DeviceDomains.join(' or '));
			}
		}

		const updateData: Record<string, unknown> = {
			device_email: deviceEmail || '',
			auto_send: autoSend === 'on'
		};
		await PUT(fetch, '/v1/devices', updateData, locals.auth);

		const updatedProfile = await GET(fetch, '/v1/user/profile', locals.auth);
		await setUserCookie(cookies, {
			account: updatedProfile.account,
			email: updatedProfile.email,
			deviceEmail: updatedProfile.device_email,
			autoSend: updatedProfile.auto_send
		});
		redirect(303, '/account');
	},
	deleteDevice: async ({ locals, cookies }) => {
		await DELETE(fetch, '/v1/devices', locals.auth);
		const profile = await GET(fetch, '/v1/user/profile', locals.auth);
		await setUserCookie(cookies, {
			account: profile.account,
			email: profile.email,
			deviceEmail: profile.device_email,
			autoSend: profile.auto_send
		});
		redirect(303, '/account');
	}
} satisfies Actions;
