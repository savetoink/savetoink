import { redirect, error, fail } from '@sveltejs/kit';
import {
	getArticle,
	favoriteArticle,
	sendArticle,
	deleteArticle,
	setTags,
	withActionFail
} from '$lib/server/apiClient';
import type { Actions, PageServerLoad, RequestEvent } from './$types';
import { ApiError } from '@savetoink/shared';

export const load: PageServerLoad = async ({
	locals,
	fetch,
	params,
	request,
	getClientAddress
}) => {
	try {
		const article = await getArticle(
			{ locals, fetch, request, getClientAddress } as RequestEvent,
			params.id
		);
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
	favorite: async ({ locals, fetch, params, request, getClientAddress }) => {
		return withActionFail(() =>
			favoriteArticle({ locals, fetch, request, getClientAddress } as RequestEvent, params.id)
		);
	},
	send: async ({ locals, fetch, params, request, getClientAddress }) => {
		return withActionFail(() =>
			sendArticle({ locals, fetch, request, getClientAddress } as RequestEvent, params.id)
		);
	},
	delete: async ({ locals, fetch, params, request, getClientAddress }) => {
		const result = await withActionFail(() =>
			deleteArticle({ locals, fetch, request, getClientAddress } as RequestEvent, params.id)
		);
		if (result && 'status' in result) {
			return result;
		}
		redirect(303, '/');
	},
	setTags: async ({ locals, fetch, params, request, getClientAddress }) => {
		const data = await request.formData();
		const tagsStr = data.get('tags');

		if (typeof tagsStr !== 'string') {
			return fail(400, { message: 'Invalid tags format' });
		}

		try {
			const tags = JSON.parse(tagsStr);
			if (!Array.isArray(tags) || !tags.every((t) => typeof t === 'string')) {
				return fail(400, { message: 'Tags must be an array of strings' });
			}

			return withActionFail(() =>
				setTags({ locals, fetch, request, getClientAddress } as RequestEvent, params.id, tags)
			);
		} catch {
			return fail(400, { message: 'Invalid JSON format' });
		}
	}
} satisfies Actions;
