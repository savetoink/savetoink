import { PUBLIC_IS_DEV_WORKER } from '$env/static/public';
import { getIsDev } from '@savetoink/shared';

const unauthenticatedPaths = ['/account', '/auth/callback', '/sentry/tunnel'];
export const isAuthenticatedPath = (path: string) => {
	try {
		const url = new URL(path, 'http://localhost');
		let pathname = url.pathname;
		if (pathname.endsWith('/')) {
			pathname = pathname.slice(0, -1);
		}
		return !unauthenticatedPaths.includes(pathname);
	} catch {
		return !unauthenticatedPaths.includes(path);
	}
};

export const isDev: boolean = getIsDev(import.meta.env.DEV, PUBLIC_IS_DEV_WORKER);
