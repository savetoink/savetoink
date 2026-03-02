import { redirect } from '@sveltejs/kit';
import { GET, POST, DELETE, PUT } from '$lib/server/apiClient';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, fetch, params }) => {
	const article = await GET(fetch, `/v1/articles/${params.id}`, locals.auth);
	return { ...article, user: locals.user };
};

export const actions = {
	favorite: async ({ locals, fetch, params }) => {
		await PUT(fetch, `/v1/articles/${params.id}/favorite`, null, locals.auth);
	},
	send: async ({ locals, fetch, params }) => {
		await POST(fetch, `/v1/articles/${params.id}/send`, null, locals.auth);
	},
	delete: async ({ locals, fetch, params }) => {
		await DELETE(fetch, `/v1/articles/${params.id}`, locals.auth);
		redirect(303, '/');
	}
} satisfies Actions;
