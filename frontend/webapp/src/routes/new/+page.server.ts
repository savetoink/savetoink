import { fail, isHttpError } from '@sveltejs/kit';
import { createArticle } from '$lib/server/apiClient';
import type { Actions, PageServerLoad, RequestEvent } from './$types';
import type { UserProfile } from '@savetoink/shared';

export const load: PageServerLoad = async ({ url, locals }) => {
	const incomingUrl = url.searchParams.get('url') ?? url.searchParams.get('text') ?? null;

	return {
		user: locals.user as UserProfile | undefined,
		incomingUrl
	};
};

export const actions: Actions = {
	new: async ({ locals, request, fetch, getClientAddress }) => {
		const data = await request.formData();
		const txt = data.get('url');
		const sendToDevice = data.get('sendToDevice');

		if (!txt || typeof txt !== 'string') {
			return fail(400, { error: 'URL is required' });
		}

		try {
			const article = await createArticle(
				{ locals, request, fetch, getClientAddress } as RequestEvent,
				txt,
				sendToDevice === 'on'
			);

			return { success: true, article };
		} catch (e) {
			let status = 500;
			let message = '';

			if (isHttpError(e)) {
				status = e.status;
				message = e.body.message;
			} else if (e && typeof e === 'object' && 'status' in e && typeof e.status === 'number') {
				status = e.status;
				if ('message' in e && typeof e.message === 'string') {
					message = e.message;
				} else if (e instanceof Error) {
					message = e.message;
				}
			} else if (e instanceof Error) {
				message = e.message;
			} else {
				message = String(e);
			}

			return fail(status, { error: 'Failed to create article: ' + message });
		}
	}
} satisfies Actions;
