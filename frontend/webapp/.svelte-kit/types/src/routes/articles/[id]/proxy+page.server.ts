// @ts-nocheck
import { redirect, error, fail } from '@sveltejs/kit';
import {
	getArticle,
	favoriteArticle,
	sendArticle,
	deleteArticle,
	addTags,
	removeTags,
	withActionFail
} from '$lib/server/apiClient';
import type { Actions, PageServerLoad, RequestEvent } from './$types';
import { ApiError } from '@savetoink/shared';

export const load = async ({
	locals,
	fetch,
	params,
	request,
	getClientAddress
}: Parameters<PageServerLoad>[0]) => {
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
	addTags: async ({ locals, fetch, params, request, getClientAddress }) => {
		const formData = await request.formData();
		const tagsStr = formData.get('tags') as string;
		if (!tagsStr) {
			return fail(400, { message: 'Tags are required' });
		}
		const tags = tagsStr
			.split(',')
			.map((t) => t.trim())
			.filter((t) => t.length > 0);
		return withActionFail(() =>
			addTags({ locals, fetch, request, getClientAddress } as RequestEvent, params.id, tags)
		);
	},
	removeTags: async ({ locals, fetch, params, request, getClientAddress }) => {
		const formData = await request.formData();
		const tagsStr = formData.get('tags') as string;
		if (!tagsStr) {
			return fail(400, { message: 'Tags are required' });
		}
		const tags = tagsStr
			.split(',')
			.map((t) => t.trim())
			.filter((t) => t.length > 0);
		return withActionFail(() =>
			removeTags({ locals, fetch, request, getClientAddress } as RequestEvent, params.id, tags)
		);
	}
} satisfies Actions;
