import { PUBLIC_API_URL } from '$env/static/public';
import { error, fail } from '@sveltejs/kit';
import { createApiClient, ApiError } from '@savetoink/shared';
import type { ApiClientOptions } from '@savetoink/shared';

function withSvelteKitError<T>(fn: () => Promise<T>): Promise<T> {
	try {
		return fn();
	} catch (e) {
		if (e instanceof ApiError) {
			error(e.status, e.message);
		}
		error(500, e instanceof Error ? e.message : 'Unknown error');
	}
}

export async function withActionFail<T>(
	fn: () => Promise<T>
): Promise<ReturnType<typeof fail> | T> {
	try {
		return await fn();
	} catch (e) {
		if (e instanceof ApiError) {
			return fail(e.status, { message: e.message });
		}
		throw e;
	}
}

function getUserAgent(request: Request): string | undefined {
	return request.headers.get('User-Agent') || undefined;
}

function createSvelteKitClient(fetch: typeof globalThis.fetch, userAgent?: string) {
	return createApiClient({
		baseUrl: PUBLIC_API_URL,
		fetch,
		userAgent
	} as ApiClientOptions);
}

function createApiClientWithUserAgent(fetch: typeof globalThis.fetch, request?: Request) {
	const userAgent = request ? getUserAgent(request) : undefined;
	return createSvelteKitClient(fetch, userAgent);
}

// Specific API methods with SvelteKit error handling
export function getProfile(fetch: typeof globalThis.fetch, token: string, request?: Request) {
	const client = createApiClientWithUserAgent(fetch, request);
	return withSvelteKitError(() => client.getProfile(token));
}

export function getArticles(
	fetch: typeof globalThis.fetch,
	params: { page?: number; page_size?: number; favorite?: boolean },
	token: string,
	request?: Request
) {
	const client = createApiClientWithUserAgent(fetch, request);
	return withSvelteKitError(() => client.getArticles(params, token));
}

export function getArticle(
	fetch: typeof globalThis.fetch,
	id: string,
	token: string,
	request?: Request
) {
	const client = createApiClientWithUserAgent(fetch, request);
	return withSvelteKitError(() => client.getArticle(id, token));
}

export function createArticle(
	fetch: typeof globalThis.fetch,
	url: string,
	sendToDevice: boolean,
	token: string,
	request?: Request
) {
	const client = createApiClientWithUserAgent(fetch, request);
	return withSvelteKitError(() => client.createArticle(url, sendToDevice, token));
}

export function sendArticle(
	fetch: typeof globalThis.fetch,
	id: string,
	token: string,
	request?: Request
) {
	const client = createApiClientWithUserAgent(fetch, request);
	return withSvelteKitError(() => client.sendArticle(id, token));
}

export function favoriteArticle(
	fetch: typeof globalThis.fetch,
	id: string,
	token: string,
	request?: Request
) {
	const client = createApiClientWithUserAgent(fetch, request);
	return withSvelteKitError(() => client.favoriteArticle(id, token));
}

export function deleteArticle(
	fetch: typeof globalThis.fetch,
	id: string,
	token: string,
	request?: Request
) {
	const client = createApiClientWithUserAgent(fetch, request);
	return withSvelteKitError(() => client.deleteArticle(id, token));
}

export function updateDevice(
	fetch: typeof globalThis.fetch,
	deviceEmail: string,
	autoSend: boolean,
	token: string,
	request?: Request
) {
	const client = createApiClientWithUserAgent(fetch, request);
	return withSvelteKitError(() => client.updateDevice(deviceEmail, autoSend, token));
}

export function deleteDevice(fetch: typeof globalThis.fetch, token: string, request?: Request) {
	const client = createApiClientWithUserAgent(fetch, request);
	return withSvelteKitError(() => client.deleteDevice(token));
}

export const exchangeCodeForToken = (
	fetch: typeof globalThis.fetch,
	code: string,
	redirectUri: string
) => withSvelteKitError(() => createSvelteKitClient(fetch).exchangeCodeForToken(code, redirectUri));
