import { error, redirect } from '@sveltejs/kit';
import { POST } from '$lib/server/apiClient';
import type { Actions } from './$types';

export const actions: Actions = {
	new: async ({ locals, request }) => {
		const data = await request.formData();
		const txt = data.get('url');
		const tagsStr = data.get('tags');

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

		await POST(fetch, `/v1/articles`, { url: txt, tags }, locals.jwt);
		redirect(303, '/');
	}
} satisfies Actions;
