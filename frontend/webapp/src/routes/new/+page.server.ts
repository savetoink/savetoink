import { error, redirect } from '@sveltejs/kit';
import { createArticle } from '$lib/server/apiClient';
import type { Actions } from './$types';

export const actions: Actions = {
	new: async ({ locals, request, fetch }) => {
		const data = await request.formData();
		const txt = data.get('url');

		if (!txt || typeof txt !== 'string') {
			error(400, 'URL is required');
		}

		await createArticle(fetch, txt, locals.auth ?? '');

		redirect(303, '/');
	}
} satisfies Actions;
