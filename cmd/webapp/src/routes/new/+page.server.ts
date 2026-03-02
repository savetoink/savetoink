import { error, redirect } from '@sveltejs/kit';
import { POST } from '$lib/server/apiClient';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	return { user: locals.user };
};

export const actions: Actions = {
	new: async ({ locals, request, fetch }) => {
		const data = await request.formData();
		const txt = data.get('url');
		const tagsStr = data.get('tags');
		const sendToDevice = data.get('sendToDevice');

		if (!txt) {
			error(400, 'URL is required');
		}

		let tags: string[] | undefined;
		if (tagsStr && typeof tagsStr === 'string') {
			tags = tagsStr
				.split(',')
				.map((tag) => tag.trim())
				.filter((tag) => tag.length > 0);
		}

		const article = await POST(fetch, `/v1/articles`, { url: txt, tags }, locals.auth);

		if (sendToDevice === 'on') {
			await POST(fetch, `/v1/articles/${article.id}/send`, {}, locals.auth);
		}

		redirect(303, '/');
	}
} satisfies Actions;
