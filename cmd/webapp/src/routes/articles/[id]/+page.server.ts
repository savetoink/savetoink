import { redirect } from '@sveltejs/kit';
import { GET, DELETE } from '$lib/server/apiClient';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, fetch, params }) => {
	return await GET(fetch, `/v1/articles/${params.id}`, locals.jwt);
};

export const actions = {
	delete: async ({ locals, fetch, params }) => {
		await DELETE(fetch, `/v1/articles/${params.id}`, locals.jwt);
		redirect(303, '/');
	}
} satisfies Actions;
