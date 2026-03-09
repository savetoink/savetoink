import { error, redirect } from '@sveltejs/kit';
import { createArticle } from '$lib/server/apiClient';
import type { Actions, PageServerLoad } from './$types';
import type { UserProfile } from '@savetoink/shared';

export const load: PageServerLoad = async ({ locals }) => {
	return { user: locals.user as UserProfile | undefined };
};

export const actions: Actions = {
	new: async ({ locals, request, fetch }) => {
		const data = await request.formData();
		const txt = data.get('url');
		const sendToDevice = data.get('sendToDevice');

		if (!txt || typeof txt !== 'string') {
			error(400, 'URL is required');
		}

		await createArticle(fetch, txt, sendToDevice === 'on', locals.auth ?? '');

		redirect(303, '/');
	}
} satisfies Actions;
