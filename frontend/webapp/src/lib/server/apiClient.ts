import { PUBLIC_API_URL } from '$env/static/public';
import { error } from '@sveltejs/kit';
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

// Helper to create API client with SvelteKit fetch
function createSvelteKitClient(fetch: typeof globalThis.fetch) {
	return createApiClient({
		baseUrl: PUBLIC_API_URL,
		fetch
	} as ApiClientOptions);
}

// Specific API methods with SvelteKit error handling
export const getProfile = (fetch: typeof globalThis.fetch, token: string) =>
	withSvelteKitError(() => createSvelteKitClient(fetch).getProfile(token));

export const getArticles = (
	fetch: typeof globalThis.fetch,
	params: { page?: number; pageSize?: number; favorite?: boolean },
	token: string
) => withSvelteKitError(() => createSvelteKitClient(fetch).getArticles(params, token));

export const getArticle = (fetch: typeof globalThis.fetch, id: string, token: string) =>
	withSvelteKitError(() => createSvelteKitClient(fetch).getArticle(id, token));

export const createArticle = (fetch: typeof globalThis.fetch, url: string, token: string) =>
	withSvelteKitError(() => createSvelteKitClient(fetch).createArticle(url, token));

export const sendArticle = (fetch: typeof globalThis.fetch, id: string, token: string) =>
	withSvelteKitError(() => createSvelteKitClient(fetch).sendArticle(id, token));

export const favoriteArticle = (fetch: typeof globalThis.fetch, id: string, token: string) =>
	withSvelteKitError(() => createSvelteKitClient(fetch).favoriteArticle(id, token));

export const deleteArticle = (fetch: typeof globalThis.fetch, id: string, token: string) =>
	withSvelteKitError(() => createSvelteKitClient(fetch).deleteArticle(id, token));

export const updateDevice = (
	fetch: typeof globalThis.fetch,
	deviceEmail: string,
	autoSend: boolean,
	token: string
) =>
	withSvelteKitError(() => createSvelteKitClient(fetch).updateDevice(deviceEmail, autoSend, token));

export const deleteDevice = (fetch: typeof globalThis.fetch, token: string) =>
	withSvelteKitError(() => createSvelteKitClient(fetch).deleteDevice(token));

export const exchangeCodeForToken = (
	fetch: typeof globalThis.fetch,
	code: string,
	redirectUri: string
) => withSvelteKitError(() => createSvelteKitClient(fetch).exchangeCodeForToken(code, redirectUri));
