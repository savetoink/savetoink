import { redirect, error } from '@sveltejs/kit';
import {
	getArticle,
	favoriteArticle,
	sendArticle,
	deleteArticle,
	withActionFail
} from '$lib/server/apiClient';
import type { Actions, PageServerLoad } from './$types';
import { ApiError } from '@savetoink/shared';

export const load: PageServerLoad = async ({ locals, fetch, params }) => {
	try {
		const article = await getArticle(fetch, params.id, locals.auth ?? '');
		return { ...article, user: locals.user };
	} catch (err) {
		if (err instanceof ApiError) {
			error(err.status, err.message);
		} else {
			throw err;
		}
	}
};

export const actions = {
	favorite: async ({ locals, fetch, params }) => {
		return withActionFail(() => favoriteArticle(fetch, params.id, locals.auth ?? ''));
	},
	send: async ({ locals, fetch, params }) => {
		return withActionFail(() => sendArticle(fetch, params.id, locals.auth ?? ''));
	},
	delete: async ({ locals, fetch, params }) => {
		const result = await withActionFail(() => deleteArticle(fetch, params.id, locals.auth ?? ''));
		if (result && 'status' in result) {
			return result;
		}
		redirect(303, '/');
	}
} satisfies Actions;
