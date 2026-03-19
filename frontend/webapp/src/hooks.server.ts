import { sequence } from '@sveltejs/kit/hooks';
import { handleErrorWithSentry, sentryHandle } from '@sentry/sveltekit';
import { handle as rootHandle } from '$lib/server/hooks';
import type { Handle } from '@sveltejs/kit';

export const handle: Handle = sequence(sentryHandle(), rootHandle);

const sentryHandleError = handleErrorWithSentry();

export const handleError = (input: { event: unknown; error: unknown; status: number }) => {
	const message = input.error instanceof Error ? input.error.message : String(input.error);
	// @ts-expect-error - Sentry's handleErrorWithSentry has complex types that don't match our simplified input
	sentryHandleError(input);
	return { message };
};
