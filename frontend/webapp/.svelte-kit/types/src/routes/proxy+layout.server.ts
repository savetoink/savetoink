// @ts-nocheck
import type { LayoutServerLoad } from './$types';

export const load = async ({ locals }: Parameters<LayoutServerLoad>[0]) => {
	return { auth: locals.auth, isLoggedIn: locals.isLoggedIn, user: locals.user };
};
