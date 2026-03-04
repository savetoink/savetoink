import { error, fail, redirect } from '@sveltejs/kit';
import {
	AUTH_KEY,
	setAuthCookie,
	deleteAuthCookie,
	setUserCookie,
	deleteUserCookie
} from '$lib/server/cookies';
import { getProfile, updateDevice, deleteDevice } from '$lib/server/apiClient';
import { DeviceDomains } from '$lib/consts';
import type { UserProfile } from '@savetoink/shared';
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

		let profile: UserProfile;
		try {
			profile = await getProfile(fetch, token);
		} catch (err) {
			console.error('failed to validate api key', err);
			return fail(400, { error: 'Invalid API key' });
		}

		setAuthCookie(cookies, token, { trim: true });
		await setUserCookie(cookies, {
			account: profile.account,
			email: profile.email,
			deviceEmail: profile.deviceEmail,
			autoSend: profile.autoSend
		});

		redirect(303, '/');
	},
	clean: async ({ cookies }) => {
		deleteAuthCookie(cookies);
		deleteUserCookie(cookies);
		redirect(303, '/account');
	},
	updateAutoSend: async ({ locals, request, fetch, cookies }) => {
		const data = await request.formData();
		const autoSend = data.get('autoSend');

		await updateDevice(fetch, locals.user?.deviceEmail || '', autoSend === 'on', locals.auth ?? '');

		const updatedProfile = await getProfile(fetch, locals.auth ?? '');
		await setUserCookie(cookies, {
			account: updatedProfile.account,
			email: updatedProfile.email,
			deviceEmail: updatedProfile.deviceEmail,
			autoSend: updatedProfile.autoSend
		});
		return { success: true };
	},
	updateProfile: async ({ locals, request, fetch, cookies }) => {
		const data = await request.formData();
		const deviceEmail = data.get('deviceEmail');
		const autoSend = data.get('autoSend');

		if (typeof deviceEmail === 'string' && deviceEmail.trim() !== '') {
			const parts = deviceEmail.split('@');
			if (parts.length !== 2 || !parts[1]) {
				error(400, 'invalid email format');
			}

			const domain = '@' + parts[1];
			if (!DeviceDomains.includes(domain)) {
				error(400, 'kindle email domain must be ' + DeviceDomains.join(' or '));
			}
		}

		await updateDevice(
			fetch,
			deviceEmail ? deviceEmail.toString() : '',
			autoSend === 'on',
			locals.auth ?? ''
		);

		const updatedProfile = await getProfile(fetch, locals.auth ?? '');
		await setUserCookie(cookies, {
			account: updatedProfile.account,
			email: updatedProfile.email,
			deviceEmail: updatedProfile.deviceEmail,
			autoSend: updatedProfile.autoSend
		});
		redirect(303, '/account');
	},
	deleteDevice: async ({ locals, fetch, cookies }) => {
		await deleteDevice(fetch, locals.auth ?? '');
		const profile = await getProfile(fetch, locals.auth ?? '');
		await setUserCookie(cookies, {
			account: profile.account,
			email: profile.email,
			deviceEmail: profile.deviceEmail,
			autoSend: profile.autoSend
		});
		redirect(303, '/account');
	}
} satisfies Actions;
