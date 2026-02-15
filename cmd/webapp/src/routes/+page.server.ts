import { fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { requireApiKey } from '$lib/server/auth';
import { ApiError } from '$lib/server/apiClient';

export const load: PageServerLoad = async ({ locals, fetch, url }) => {
	const apiClient = requireApiKey(locals);

	const pageParam = url.searchParams.get('page');
	const pageSizeParam = url.searchParams.get('page_size');

	const page = pageParam ? parseInt(pageParam, 10) : 1;
	const pageSize = pageSizeParam ? parseInt(pageSizeParam, 10) : 10;

	const response = await apiClient.getArticles(fetch, page, pageSize);
	return {
		articles: response.articles,
		total: response.total,
		page: response.page,
		page_size: response.page_size,
		has_more: response.has_more
	};
};

export const actions: Actions = {
	delete: async ({ locals, fetch, request }) => {
		const apiClient = requireApiKey(locals);

		const formData = await request.formData();
		const id = formData.get('id');

		if (!id || typeof id !== 'string') {
			return fail(400, { error: 'article id is required' });
		}

		try {
			await apiClient.deleteArticle(id, fetch);
			return { success: true };
		} catch (err) {
			const status = err instanceof ApiError ? err.status : 500;
			const message = err instanceof Error ? err.message : 'failed to delete article';
			return fail(status, { error: message });
		}
	}
};
