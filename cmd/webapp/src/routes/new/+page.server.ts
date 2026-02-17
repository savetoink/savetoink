import type { Actions } from './$types';
import { redirect } from '@sveltejs/kit';
import { POST } from '$lib/server/apiClient';

export const load = async ({ locals }) => {
	if (!locals.jwt) {
		throw redirect(302, '/login');
	}
};

export const actions = {
	new: async ({ locals, request }) => {
		const data = await request.formData();
		const txt = data.get('url') || 'sfda';

		try {
			await POST(fetch, `/v1/articles`, { url: txt }, locals.jwt);
			redirect(303, '/');
		} catch {
			redirect(302, '/');
		}
	}
} satisfies Actions;
