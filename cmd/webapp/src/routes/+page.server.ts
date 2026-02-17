import type { PageServerLoad } from './$types';
import { GET } from '$lib/server/apiClient';
import { redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ locals, fetch, url }) => {
	const pageParam = url.searchParams.get('page');
	const pageSizeParam = url.searchParams.get('page_size');

	const page = pageParam ? parseInt(pageParam, 10) : 1;
	const pageSize = pageSizeParam ? parseInt(pageSizeParam, 10) : 10;

	try {
		return await GET(fetch, `/v1/articles?page=${page}&page_size=${pageSize}`, locals.jwt);
	} catch {
		redirect(302, '/login');
	}
};
