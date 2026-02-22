import { PUBLIC_API_URL } from '$env/static/public';
import { error } from '@sveltejs/kit';

async function send(
	fetch: typeof globalThis.fetch,
	method: string,
	path: string,
	data?: unknown,
	token?: string
) {
	const headers: Record<string, string> = {};
	if (data) {
		headers['Content-Type'] = 'application/json';
	}

	if (token) {
		headers['Authorization'] = `Bearer ${token}`;
	}

	const body = data ? JSON.stringify(data) : undefined;
	const opts = { method, headers, body };

	const res = await fetch(`${PUBLIC_API_URL}${path}`, opts);
	if (res.ok) {
		const text = await res.text();
		return text ? JSON.parse(text) : {};
	}

	const text = await res.text();
	const errorBody = text ? JSON.parse(text) : {};
	error(res.status, errorBody.error || errorBody.message || text);
}

export async function GET(fetch: typeof globalThis.fetch, path: string, token?: string) {
	return send(fetch, 'GET', path, undefined, token);
}

export async function DELETE(fetch: typeof globalThis.fetch, path: string, token?: string) {
	return send(fetch, 'DELETE', path, undefined, token);
}

export async function POST(
	fetch: typeof globalThis.fetch,
	path: string,
	data?: unknown,
	token?: string
) {
	return send(fetch, 'POST', path, data, token);
}

export async function PUT(
	fetch: typeof globalThis.fetch,
	path: string,
	data?: unknown,
	token?: string
) {
	return send(fetch, 'PUT', path, data, token);
}
