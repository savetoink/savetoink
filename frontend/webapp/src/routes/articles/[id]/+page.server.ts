import { redirect } from '@sveltejs/kit';
import { getArticle, favoriteArticle, sendArticle, deleteArticle } from '$lib/server/apiClient';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, fetch, params }) => {
	const article = await getArticle(fetch, params.id, locals.auth ?? '');
	return { ...article, user: locals.user };
};

export const actions = {
	favorite: async ({ locals, fetch, params }) => {
		await favoriteArticle(fetch, params.id, locals.auth ?? '');
	},
	send: async ({ locals, fetch, params }) => {
		await sendArticle(fetch, params.id, locals.auth ?? '');
	},
	delete: async ({ locals, fetch, params }) => {
		await deleteArticle(fetch, params.id, locals.auth ?? '');
		redirect(303, '/');
	}
} satisfies Actions;
