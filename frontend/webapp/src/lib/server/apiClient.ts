import { env as publicEnv } from '$env/dynamic/public';
import { error, fail } from '@sveltejs/kit';
import { createApiClient as createBaseApiClient, ApiError } from '@savetoink/shared';
import type { ApiClientOptions } from '@savetoink/shared';
import type { RequestEvent } from '@sveltejs/kit';

async function withSvelteKitError<T>(fn: () => Promise<T>): Promise<T> {
	try {
		return await fn();
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

function createApiClient(event: RequestEvent) {
	const clientAddress = event.getClientAddress();
	const existingXff = event.request.headers.get('x-forwarded-for');
	const xForwardedFor = clientAddress
		? existingXff
			? `${existingXff}, ${clientAddress}`
			: clientAddress
		: undefined;

	const cloudFlareRay = event.request.headers.get('cf-ray');

	return createBaseApiClient({
		baseUrl: publicEnv.PUBLIC_API_URL,
		fetch: event.fetch,
		userAgent: event.request.headers.get('User-Agent') || undefined,
		xForwardedFor,
		cloudFlareRay
	} as ApiClientOptions);
}

// Specific API methods with SvelteKit error handling
export function getProfile(event: RequestEvent) {
	const client = createApiClient(event);
	return withSvelteKitError(() => client.getProfile(event.locals.auth ?? ''));
}

export function getArticles(
	event: RequestEvent,
	params: { page?: number; page_size?: number; favorite?: boolean }
) {
	const client = createApiClient(event);
	return withSvelteKitError(() => client.getArticles(params, event.locals.auth ?? ''));
}

export function getArticle(event: RequestEvent, id: string) {
	const client = createApiClient(event);
	return withSvelteKitError(() => client.getArticle(id, event.locals.auth ?? ''));
}

export function createArticle(event: RequestEvent, url: string, sendToDevice: boolean) {
	const client = createApiClient(event);
	return withSvelteKitError(() => client.createArticle(url, sendToDevice, event.locals.auth ?? ''));
}

export function sendArticle(event: RequestEvent, id: string) {
	const client = createApiClient(event);
	return withSvelteKitError(() => client.sendArticle(id, event.locals.auth ?? ''));
}

export function favoriteArticle(event: RequestEvent, id: string) {
	const client = createApiClient(event);
	return withSvelteKitError(() => client.favoriteArticle(id, event.locals.auth ?? ''));
}

export function deleteArticle(event: RequestEvent, id: string) {
	const client = createApiClient(event);
	return withSvelteKitError(() => client.deleteArticle(id, event.locals.auth ?? ''));
}

export function updateDevice(event: RequestEvent, deviceEmail: string, autoSend: boolean) {
	const client = createApiClient(event);
	return withSvelteKitError(() =>
		client.updateDevice(deviceEmail, autoSend, event.locals.auth ?? '')
	);
}

export function deleteDevice(event: RequestEvent) {
	const client = createApiClient(event);
	return withSvelteKitError(() => client.deleteDevice(event.locals.auth ?? ''));
}

export const exchangeCodeForToken = (event: RequestEvent, code: string, redirectUri: string) =>
	withSvelteKitError(() => createApiClient(event).exchangeCodeForToken(code, redirectUri));

export function getSends(event: RequestEvent) {
	const client = createApiClient(event);
	return withSvelteKitError(() => client.getSends(event.locals.auth ?? ''));
}
