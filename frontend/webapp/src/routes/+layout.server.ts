import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ locals }) => {
	return { auth: locals.auth, isLoggedIn: locals.isLoggedIn, user: locals.user };
};
