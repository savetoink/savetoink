import type { PageServerLoad, RequestEvent, Actions } from './$types';
import { getArticles, removeTags, withActionFail } from '$lib/server/apiClient';

export const load: PageServerLoad = async ({ locals, fetch, url, request, getClientAddress }) => {
	const pageParam = url.searchParams.get('page');
	const pageSizeParam = url.searchParams.get('page_size');
	const favoritesParam = url.searchParams.get('favorite');

	const page = pageParam ? parseInt(pageParam, 10) : 1;
	const pageSize = pageSizeParam ? parseInt(pageSizeParam, 10) : 10;
	const favorite = favoritesParam === 'true' ? true : undefined;
	const tag = url.searchParams.get('tag') ?? undefined;

	const data = await getArticles({ locals, fetch, request, getClientAddress } as RequestEvent, {
		page,
		page_size: pageSize,
		favorite,
		tag
	});

	return { ...data, user: locals.user, tag: tag, favorite };
};

export const actions = {
	removeTag: async ({ locals, fetch, request, getClientAddress }) => {
		const formData = await request.formData();
		const articleId = formData.get('articleId');
		const tagName = formData.get('tag');

		if (!articleId || typeof articleId !== 'string') {
			return withActionFail(() => {
				throw new Error('articleId is required');
			});
		}

		if (!tagName || typeof tagName !== 'string') {
			return withActionFail(() => {
				throw new Error('tag is required');
			});
		}

		return withActionFail(() =>
			removeTags({ locals, fetch, request, getClientAddress } as RequestEvent, articleId, [tagName])
		);
	}
} satisfies Actions;
