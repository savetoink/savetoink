import type { PageServerLoad, RequestEvent } from './$types';
import { getArticles } from '$lib/server/apiClient';

export const load: PageServerLoad = async ({ locals, fetch, url, request, getClientAddress }) => {
	const pageParam = url.searchParams.get('page');
	const pageSizeParam = url.searchParams.get('page_size');
	const favoritesParam = url.searchParams.get('favorites');
	const tagParam = url.searchParams.get('tag');

	const page = pageParam ? parseInt(pageParam, 10) : 1;
	const pageSize = pageSizeParam ? parseInt(pageSizeParam, 10) : 10;
	const favorite = favoritesParam === 'true' ? true : undefined;
	const tag = tagParam || undefined;

	const data = await getArticles({ locals, fetch, request, getClientAddress } as RequestEvent, {
		page,
		page_size: pageSize,
		favorite,
		tag
	});

	return { ...data, user: locals.user };
};
