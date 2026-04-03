// @ts-nocheck
import type { PageServerLoad, RequestEvent } from './$types';
import { getArticles } from '$lib/server/apiClient';

export const load = async ({ locals, fetch, url, request, getClientAddress }: Parameters<PageServerLoad>[0]) => {
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
