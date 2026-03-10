import { error, redirect } from '@sveltejs/kit';
import { createArticle } from '$lib/server/apiClient';
import type { Actions, PageServerLoad, RequestEvent } from './$types';
import type { UserProfile } from '@savetoink/shared';

export const load: PageServerLoad = async ({ locals }) => {
	return { user: locals.user as UserProfile | undefined };
};

export const actions: Actions = {
	new: async ({ locals, request }) => {
		const data = await request.formData();
		const txt = data.get('url');
		const sendToDevice = data.get('sendToDevice');

		if (!txt || typeof txt !== 'string') {
			error(400, 'URL is required');
		}

		await createArticle({ locals, request } as RequestEvent, txt, sendToDevice === 'on');

		redirect(303, '/articles');
	}
} satisfies Actions;
