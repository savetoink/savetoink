import { PUBLIC_IS_DEV_WORKER } from '$env/static/public';

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

type authBackend = 'auth0' | 'sharedKey';
export const Auth0: authBackend = 'auth0';
export const SharedKey: authBackend = 'sharedKey';

export const DeviceDomains = [
	'@kindle.com',
	'@free.kindle.com',
	'@send.kobo.com',
	'@pbsync.com',
	'@mytolino.com'
];

export const isDev: boolean = import.meta.env.DEV || PUBLIC_IS_DEV_WORKER === 'true';
