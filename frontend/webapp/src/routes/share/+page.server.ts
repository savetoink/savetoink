import { error, redirect } from '@sveltejs/kit';
import { createArticle } from '$lib/server/apiClient';
import type { Actions, PageServerLoad, RequestEvent } from './$types';
import type { UserProfile } from '@savetoink/shared';

export const load: PageServerLoad = async ({ url, locals }) => {
	return {
		incomingUrl: url.searchParams.get('url') ?? url.searchParams.get('text') ?? null,
		user: locals.user as UserProfile | undefined
	};
};

export const actions: Actions = {
	save: async ({ locals, request, fetch, getClientAddress }) => {
		const data = await request.formData();
		const incomingUrl = data.get('url');

		if (!incomingUrl || typeof incomingUrl !== 'string') {
			error(400, 'URL is required');
		}

		await createArticle(
			{ locals, request, fetch, getClientAddress } as RequestEvent,
			incomingUrl,
			false // sendToDevice — could expose this later
		);

		redirect(303, '/articles');
	}
};
