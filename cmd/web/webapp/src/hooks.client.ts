import { handleErrorWithSentry } from '@sentry/sveltekit';
import * as Sentry from '@sentry/sveltekit';
import { PUBLIC_SENTRY_DSN } from '$env/static/public';
import { isDev } from '$lib/consts';
import { getCSS } from '@savetoink/shared/css';

getCSS(isDev);

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
