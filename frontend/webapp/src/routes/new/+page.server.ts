import { error, redirect } from '@sveltejs/kit';
import { createArticle } from '$lib/server/apiClient';
import type { Actions, PageServerLoad, RequestEvent } from './$types';
import type { UserProfile } from '@savetoink/shared';

export const load: PageServerLoad = async ({ url, locals, fetch, getClientAddress }) => {
	const incomingUrl = url.searchParams.get('url') ?? url.searchParams.get('text') ?? null;

	if (incomingUrl) {
		const user = locals.user as UserProfile | undefined;
		const sendToDevice = user?.auto_send ?? false;

		const clientAddress = getClientAddress();
		const auth = locals.auth ?? '';

		await createArticle(
			{
				request: { headers: new Headers() } as Request,
				fetch,
				getClientAddress: () => clientAddress,
				locals: { auth }
			} as RequestEvent,
			incomingUrl,
			sendToDevice
		);

		redirect(303, '/articles');
	}

	return {
		user: locals.user as UserProfile | undefined
	};
};

export const actions: Actions = {
	new: async ({ locals, request, fetch, getClientAddress }) => {
		const data = await request.formData();
		const txt = data.get('url');
		const sendToDevice = data.get('sendToDevice');

		if (!txt || typeof txt !== 'string') {
			error(400, 'URL is required');
		}

		await createArticle(
			{ locals, request, fetch, getClientAddress } as RequestEvent,
			txt,
			sendToDevice === 'on'
		);

		redirect(303, '/articles');
	}
} satisfies Actions;
