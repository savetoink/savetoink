import { handleErrorWithSentry } from '@sentry/sveltekit';
import * as Sentry from '@sentry/sveltekit';
import { PUBLIC_SENTRY_DSN, PUBLIC_SENTRY_ENVIRONMENT } from '$env/static/public';
import { getVersionText } from '$lib/appUtils';

// Load Temporal polyfill only if not natively supported
if (typeof globalThis.Temporal === 'undefined') {
	await import('temporal-polyfill/global');
}

if (!import.meta.env.DEV) {
	Sentry.init({
		dsn: PUBLIC_SENTRY_DSN,
		environment: PUBLIC_SENTRY_ENVIRONMENT,
		release: getVersionText(),
		tracesSampleRate: 1.0,
		enableLogs: true,
		tunnel: '/sentry/tunnel',
		sendDefaultPii: true
	});
}

const sentryHandleError = handleErrorWithSentry();

export const handleError = (input: { error: unknown }) => {
	const message = input.error instanceof Error ? input.error.message : String(input.error);
	// @ts-expect-error - Sentry's handleErrorWithSentry has complex types that don't match our simplified input
	sentryHandleError(input);
	return { message };
};
