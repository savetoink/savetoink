import type { PageServerLoad } from './$types';
import { GET } from '$lib/server/apiClient';

export const load: PageServerLoad = async ({ locals, fetch, url }) => {
	const pageParam = url.searchParams.get('page');
	const pageSizeParam = url.searchParams.get('page_size');
	const favoriteParam = url.searchParams.get('favorite');

	const page = pageParam ? parseInt(pageParam, 10) : 1;
	const pageSize = pageSizeParam ? parseInt(pageSizeParam, 10) : 10;
	const favFilter = favoriteParam === 'true';

	let path = `/v1/articles?page=${page}&page_size=${pageSize}`;
	if (favFilter) {
		path += '&favorite=true';
	}

	// TODO: return typed response: https://svelte.dev/docs/kit/types
	// https://github.com/savetoink/savetoink/issues/2
	const data = await GET(fetch, path, locals.jwt);
	return { ...data, user: locals.user };
};
