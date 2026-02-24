import { redirect } from '@sveltejs/kit';
import { GET, DELETE, PUT } from '$lib/server/apiClient';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, fetch, params }) => {
	return await GET(fetch, `/v1/articles/${params.id}`, locals.jwt);
};

export const actions = {
	favorite: async ({ locals, fetch, params }) => {
		await PUT(fetch, `/v1/articles/${params.id}/favorite`, null, locals.jwt);
	},
	delete: async ({ locals, fetch, params }) => {
		await DELETE(fetch, `/v1/articles/${params.id}`, locals.jwt);
		redirect(303, '/');
	}
} satisfies Actions;
