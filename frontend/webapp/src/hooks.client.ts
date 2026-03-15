import { handleErrorWithSentry } from '@sentry/sveltekit';
import * as Sentry from '@sentry/sveltekit';
import { PUBLIC_SENTRY_DSN, PUBLIC_SENTRY_ENVIRONMENT } from '$env/static/public';

// Load Temporal polyfill only if not natively supported
if (typeof globalThis.Temporal === 'undefined') {
	await import('temporal-polyfill/global');
}

if (!import.meta.env.DEV) {
	Sentry.init({
		dsn: PUBLIC_SENTRY_DSN,
		environment: PUBLIC_SENTRY_ENVIRONMENT,
		tracesSampleRate: 1.0,
		enableLogs: true,
		tunnel: '/sentry/tunnel',
		sendDefaultPii: true
	});
}

export const handleError = handleErrorWithSentry();
