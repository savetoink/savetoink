import { redirect, fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { requireApiKey } from '$lib/server/auth';
import { ApiError } from '$lib/server/apiClient';

export const load: PageServerLoad = async ({ locals, fetch, params }) => {
	const apiClient = requireApiKey(locals);

	const id = params.id;

	const article = await apiClient.getArticle(id, fetch);
	return {
		article
	};
};

export const actions: Actions = {
	delete: async ({ locals, fetch, params }) => {
		const apiClient = requireApiKey(locals);

		const id = params.id;

		try {
			await apiClient.deleteArticle(id, fetch);
			throw redirect(303, '/');
		} catch (err) {
			if (err instanceof Response && err.status === 303) {
				throw err;
			}
			const status = err instanceof ApiError ? err.status : 500;
			const message = err instanceof Error ? err.message : 'failed to delete article';
			return fail(status, { error: message });
		}
	}
};
