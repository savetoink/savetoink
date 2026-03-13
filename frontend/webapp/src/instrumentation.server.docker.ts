import { init, sentryHandle, handleErrorWithSentry } from '@sentry/sveltekit';
import { sequence } from '@sveltejs/kit/hooks';

init({
	dsn: process.env.PUBLIC_SENTRY_DSN,
	environment: process.env.PUBLIC_SENTRY_ENVIRONMENT,
	tracesSampleRate: 1.0
});

export const handle = sequence(sentryHandle());
export const handleError = handleErrorWithSentry();
