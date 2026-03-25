import { getAllTags, getArticles } from '$lib/server/apiClient';
import type { PageServerLoad, RequestEvent } from './$types';

export const load: PageServerLoad = async ({ locals, fetch, request, getClientAddress }) => {
	// Get all unique tags for the user
	const allTags = await getAllTags({ locals, fetch, request, getClientAddress } as RequestEvent);

	// Get article counts for each tag
	const tagsWithCounts = await Promise.all(
		allTags.tags.map(async (tag) => {
			const result = await getArticles(
				{ locals, fetch, request, getClientAddress } as RequestEvent,
				{ tag, page: 1, page_size: 1 }
			);
			return { tag, count: result.total };
		})
	);

	// Sort by article count (descending), then alphabetically
	const sortedTags = tagsWithCounts.sort((a, b) => {
		if (b.count !== a.count) {
			return b.count - a.count;
		}
		return a.tag.localeCompare(b.tag);
	});

	return { tags: sortedTags, user: locals.user };
};
