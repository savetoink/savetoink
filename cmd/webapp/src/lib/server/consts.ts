const unauthenticatedPaths = ['/login', '/auth/callback'];

export const isAuthenticatedPath = (path: string) => !unauthenticatedPaths.includes(path);
