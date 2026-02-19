import { fail, redirect } from '@sveltejs/kit';
import { JWT_KEY, setJwtCookie, deleteJwtCookie } from '$lib/server/cookies';
import { GET, PUT } from '$lib/server/apiClient';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch, locals }) => {
	if (locals.jwt) {
		return await GET(fetch, '/v1/user/profile', locals.jwt);
	}
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
	profile: async ({ locals, request }) => {
		const data = await request.formData();
		const kindleEmail = data.get('kindleEmail');

		if (!kindleEmail || typeof kindleEmail !== 'string' || kindleEmail.trim() === '') {
			return fail(400, { error: 'kindle email is required' });
		}

		await PUT(
			fetch,
			'/v1/user/profile',
			{
				kindle_email: kindleEmail
			},
			locals.jwt
		);
	}
} satisfies Actions;
