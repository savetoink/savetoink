import { initCloudflareSentryHandle } from '@sentry/sveltekit';
import { PUBLIC_SENTRY_DSN, PUBLIC_SENTRY_ENVIRONMENT } from '$env/static/public';
import { dev } from '$app/environment';

if (!dev) {
	// Notice that we support separate environment variables for deployed dev and prod
	// this runs for non-local environments only
	initCloudflareSentryHandle({
		dsn: PUBLIC_SENTRY_DSN,
		environment: PUBLIC_SENTRY_ENVIRONMENT,
		tracesSampleRate: 1.0,
		enableLogs: true
	});
}
