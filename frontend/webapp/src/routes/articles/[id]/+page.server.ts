import { redirect, error } from '@sveltejs/kit';
import {
	getArticle,
	favoriteArticle,
	sendArticle,
	deleteArticle,
	removeTags,
	addTags,
	withActionFail
} from '$lib/server/apiClient';
import type { Actions, PageServerLoad, RequestEvent } from './$types';
import { ApiError, MAX_TAGS, MAX_TAG_LENGTH } from '@savetoink/shared';

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
	removeTag: async ({ locals, fetch, params, request, getClientAddress }) => {
		const formData = await request.formData();
		const tagName = formData.get('tag');

		if (!tagName || typeof tagName !== 'string') {
			return withActionFail(() => {
				throw new Error('tag is required');
			});
		}

		return withActionFail(() =>
			removeTags({ locals, fetch, request, getClientAddress } as RequestEvent, params.id, [tagName])
		);
	},
	addTags: async ({ locals, fetch, params, request, getClientAddress }) => {
		const formData = await request.formData();
		const tagsStr = formData.get('tags');

		if (!tagsStr || typeof tagsStr !== 'string') {
			return withActionFail(() => {
				throw new Error('tags is required');
			});
		}

		// Parse tags from comma-separated string
		const tags = tagsStr
			.split(',')
			.map((tag) => tag.trim())
			.filter((tag) => tag.length > 0);

		if (tags.length === 0) {
			return withActionFail(() => {
				throw new Error('at least one tag is required');
			});
		}

		// Validate individual tag length
		for (const tag of tags) {
			if (tag.length > MAX_TAG_LENGTH) {
				return withActionFail(() => {
					throw new Error(`tag "${tag}" exceeds maximum length of ${MAX_TAG_LENGTH} characters`);
				});
			}
		}

		// Fetch current article to validate total tag count
		const currentArticle = await getArticle(
			{ locals, fetch, request, getClientAddress } as RequestEvent,
			params.id
		);

		const currentTagCount = currentArticle.tags?.length || 0;
		const totalTags = currentTagCount + tags.length;

		if (totalTags > MAX_TAGS) {
			return withActionFail(() => {
				throw new Error(`maximum ${MAX_TAGS} tags allowed per article`);
			});
		}

		return withActionFail(() =>
			addTags({ locals, fetch, request, getClientAddress } as RequestEvent, params.id, tags)
		);
	}
} satisfies Actions;
