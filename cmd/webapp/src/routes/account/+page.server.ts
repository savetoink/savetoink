import { error, fail, redirect } from '@sveltejs/kit';
import { JWT_KEY, setJwtCookie, deleteJwtCookie } from '$lib/server/cookies';
import { GET, PUT, DELETE } from '$lib/server/apiClient';
import { KindleDomains } from '$lib/consts';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch, locals, cookies }) => {
	if (locals.jwt) {
		try {
			return await GET(fetch, '/v1/user/profile', locals.jwt);
		} catch (err) {
			deleteJwtCookie(cookies);
			console.error('failed to load user profile', err);
		}
	}
	return {};
};

export const actions: Actions = {
	save: async ({ cookies, request }) => {
		const data = await request.formData();
		const token = data.get(JWT_KEY);

		if (!token || typeof token !== 'string' || token.trim() === '') {
			return fail(400, { error: 'api key is required' });
		}

		setJwtCookie(cookies, token, { trim: true });

		redirect(303, '/');
	},
	clean: async ({ cookies }) => {
		deleteJwtCookie(cookies);
		redirect(303, '/account');
	},
	updateProfile: async ({ locals, request }) => {
		const data = await request.formData();
		const kindleEmail = data.get('kindleEmail');

		if (!kindleEmail || typeof kindleEmail !== 'string' || kindleEmail.trim() === '') {
			return error(400, 'kindle email is required');
		}

		const parts = kindleEmail.split('@');
		if (parts.length !== 2 || !parts[1]) {
			return error(400, 'invalid email format');
		}

		const domain = '@' + parts[1];
		if (!KindleDomains.includes(domain)) {
			return error(400, 'kindle email domain must be ' + KindleDomains.join(' or '));
		}

		try {
			await PUT(
				fetch,
				'/v1/user/profile',
				{
					kindle_email: kindleEmail
				},
				locals.jwt
			);
		} catch (err) {
			error(400, `failed to update user profile: ${err}`);
		}
	},
	deleteProfile: async ({ locals }) => {
		await DELETE(fetch, '/v1/user/profile', locals.jwt);
	}
} satisfies Actions;
