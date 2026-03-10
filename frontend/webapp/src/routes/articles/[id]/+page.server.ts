import { redirect, error } from '@sveltejs/kit';
import {
	getArticle,
	favoriteArticle,
	sendArticle,
	deleteArticle,
	withActionFail
} from '$lib/server/apiClient';
import type { Actions, PageServerLoad, RequestEvent } from './$types';
import { ApiError } from '@savetoink/shared';

export const load: PageServerLoad = async ({ locals, fetch, params, request }) => {
	try {
		const article = await getArticle({ locals, fetch, request } as RequestEvent, params.id);
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
	favorite: async ({ locals, fetch, params, request }) => {
		return withActionFail(() =>
			favoriteArticle({ locals, fetch, request } as RequestEvent, params.id)
		);
	},
	send: async ({ locals, fetch, params, request }) => {
		return withActionFail(() => sendArticle({ locals, fetch, request } as RequestEvent, params.id));
	},
	delete: async ({ locals, fetch, params, request }) => {
		const result = await withActionFail(() =>
			deleteArticle({ locals, fetch, request } as RequestEvent, params.id)
		);
		if (result && 'status' in result) {
			return result;
		}
		redirect(303, '/');
	}
} satisfies Actions;
