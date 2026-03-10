import type { PageServerLoad } from './$types';
import { getArticles } from '$lib/server/apiClient';

export const load: PageServerLoad = async ({ locals, fetch, url, request }) => {
	const pageParam = url.searchParams.get('page');
	const pageSizeParam = url.searchParams.get('page_size');
	const favoriteParam = url.searchParams.get('favorite');

	const page = pageParam ? parseInt(pageParam, 10) : 1;
	const pageSize = pageSizeParam ? parseInt(pageSizeParam, 10) : 10;
	const favFilter = favoriteParam === 'true';

	const data = await getArticles(
		fetch,
		{ page, page_size: pageSize, favorite: favFilter },
		locals.auth ?? '',
		request
	);

	return { ...data, user: locals.user };
};
