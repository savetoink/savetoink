import { initCloudflareSentryHandle } from '@sentry/sveltekit';
import { PUBLIC_SENTRY_DSN } from '$env/static/public';
import { dev } from '$app/environment';

if (!dev) {
	initCloudflareSentryHandle({
		dsn: PUBLIC_SENTRY_DSN,
		tracesSampleRate: 1.0,
		enableLogs: true
	});
}
