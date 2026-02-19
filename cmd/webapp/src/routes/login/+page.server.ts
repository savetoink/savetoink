import { fail, redirect } from '@sveltejs/kit';
import { JWT_KEY, setJwtCookie, deleteJwtCookie } from '$lib/server/cookies';
import type { Actions } from './$types';

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
		redirect(303, '/login');
	}
} satisfies Actions;
