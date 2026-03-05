import { handleErrorWithSentry } from '@sentry/sveltekit';
import * as Sentry from '@sentry/sveltekit';
import { PUBLIC_SENTRY_DSN } from '$env/static/public';
import { isDev } from '$lib/consts';
import '@savetoink/shared/css';

if (isDev) {
	import('@savetoink/shared/css-dev');
}

if (!import.meta.env.DEV) {
	Sentry.init({
		dsn: PUBLIC_SENTRY_DSN,
		tracesSampleRate: 1.0,
		enableLogs: true,
		tunnel: '/sentry/tunnel',
		sendDefaultPii: true
	});
}

export const handleError = handleErrorWithSentry();
