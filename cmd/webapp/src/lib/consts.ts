const unauthenticatedPaths = ['/account', '/auth/callback'];
export const isAuthenticatedPath = (path: string) => !unauthenticatedPaths.includes(path);

type authBackend = 'auth0' | 'sharedKey';
export const Auth0: authBackend = 'auth0';
export const SharedKey: authBackend = 'sharedKey';
