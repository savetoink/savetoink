import { error, redirect } from '@sveltejs/kit';
import { POST } from '$lib/server/apiClient';
import type { Actions } from './$types';

export const actions: Actions = {
	new: async ({ locals, request }) => {
		const data = await request.formData();
		const txt = data.get('url');

		if (!txt) {
			error(400, 'URL is required');
		}

		await POST(fetch, `/v1/articles`, { url: txt }, locals.jwt);
		redirect(303, '/');
	}
} satisfies Actions;
