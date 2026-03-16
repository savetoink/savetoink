import { sequence } from '@sveltejs/kit/hooks';
import { handleErrorWithSentry, sentryHandle } from '@sentry/sveltekit';
import { handle as rootHandle } from '$lib/server/hooks';
import type { Handle } from '@sveltejs/kit';

export const handle: Handle = sequence(sentryHandle(), rootHandle);

export const handleError = handleErrorWithSentry();
