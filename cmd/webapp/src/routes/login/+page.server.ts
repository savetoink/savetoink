import { fail, redirect } from '@sveltejs/kit';
import type { Actions } from './$types';
const jwtKey = 'jwt';

export const actions: Actions = {
	save: async ({ cookies, request }) => {
		const data = await request.formData();
		const token = data.get(jwtKey);

		if (!token || typeof token !== 'string' || token.trim() === '') {
			return fail(400, { error: 'api key is required' });
		}

		cookies.set(jwtKey, token.trim(), {
			path: '/',
			httpOnly: true,
			secure: import.meta.env.PROD,
			sameSite: 'lax',
			maxAge: 60 * 60 * 24 * 365
		});

		redirect(303, '/');
	},
	clean: async ({ cookies }) => {
		cookies.delete(jwtKey, { path: '/' });
		redirect(303, '/settings');
	}
};
