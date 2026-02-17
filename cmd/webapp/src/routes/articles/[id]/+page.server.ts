import type { Actions, PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { GET, DELETE } from '$lib/server/apiClient';

export const load: PageServerLoad = async ({ locals, fetch, params }) => {
	try {
		return await GET(fetch, `/v1/articles/${params.id}`, locals.jwt);
	} catch {
		redirect(303, '/');
	}
};

export const actions = {
	delete: async ({ locals, fetch, params }) => {
		await DELETE(fetch, `/v1/articles/${params.id}`, locals.jwt);
		redirect(303, '/');
	}
} satisfies Actions;
