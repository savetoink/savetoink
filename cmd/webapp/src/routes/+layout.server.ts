import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ locals }) => {
	return { jwt: locals.jwt, isLoggedIn: locals.isLoggedIn, user: locals.user };
};
