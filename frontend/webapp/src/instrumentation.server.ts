import { initCloudflareSentryHandle } from '@sentry/sveltekit';
import { env as publicEnv } from '$env/dynamic/public';
import { dev } from '$app/environment';
import { getVersionText } from '$lib/appUtils';

if (!dev) {
	// Notice that we support separate environment variables for deployed dev and prod
	// this runs for non-local environments only
	initCloudflareSentryHandle({
		dsn: publicEnv.PUBLIC_SENTRY_DSN,
		environment: publicEnv.PUBLIC_SENTRY_ENVIRONMENT,
		release: getVersionText(),
		tracesSampleRate: 1.0,
		enableLogs: true
	});
}
