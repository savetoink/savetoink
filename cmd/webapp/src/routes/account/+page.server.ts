import { error, fail, redirect } from '@sveltejs/kit';
import { JWT_KEY, setJwtCookie, deleteJwtCookie } from '$lib/server/cookies';
import { GET, PUT, DELETE } from '$lib/server/apiClient';
import { DeviceDomains } from '$lib/consts';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	return { user: locals.user };
};

export const actions: Actions = {
	save: async ({ cookies, request, fetch }) => {
		const data = await request.formData();
		const token = data.get(JWT_KEY);

		if (!token || typeof token !== 'string' || token.trim() === '') {
			return fail(400, { error: 'api key is required' });
		}

		try {
			await GET(fetch, '/v1/user/profile', token);
		} catch (err) {
			console.error('failed to validate api key', err);
			return fail(400, { error: 'Invalid API key' });
		}

		setJwtCookie(cookies, token, { trim: true });

		redirect(303, '/');
	},
	clean: async ({ cookies }) => {
		deleteJwtCookie(cookies);
		redirect(303, '/account');
	},
	updateProfile: async ({ locals, request, fetch }) => {
		const data = await request.formData();
		const deviceEmail = data.get('deviceEmail');
		const autoSend = data.get('autoSend');

		const profile = await GET(fetch, '/v1/user/profile', locals.jwt);

		if (!deviceEmail || typeof deviceEmail !== 'string' || deviceEmail.trim() === '') {
			if (!profile.device_email) {
				return error(400, 'device email is required');
			}
		}

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
			device_email: deviceEmail || profile.device_email,
			auto_send: autoSend === 'on'
		};
		await PUT(fetch, '/v1/user/profile', updateData, locals.jwt);
	},
	deleteProfile: async ({ locals }) => {
		await DELETE(fetch, '/v1/user/profile', locals.jwt);
	}
} satisfies Actions;
