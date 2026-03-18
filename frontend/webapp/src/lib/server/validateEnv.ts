import { error } from '@sveltejs/kit';
import { env as publicEnv } from '$env/dynamic/public';
import { Auth0, SharedKey } from '@savetoink/shared';
import { PUBLIC_AUTH_BACKEND } from '$env/static/public';

export const validateEnv = () => {
	const requiredEnvVars = [
		{ name: 'PUBLIC_API_URL', value: publicEnv.PUBLIC_API_URL },
		{ name: 'PUBLIC_AUTH_BACKEND', value: PUBLIC_AUTH_BACKEND }
	];

	if (PUBLIC_AUTH_BACKEND === Auth0) {
		requiredEnvVars.push(
			{ name: 'PUBLIC_AUTH0_CLIENT_ID', value: publicEnv.PUBLIC_AUTH0_CLIENT_ID },
			{ name: 'PUBLIC_AUTH0_DOMAIN', value: publicEnv.PUBLIC_AUTH0_DOMAIN },
			{ name: 'PUBLIC_AUTH0_AUDIENCE', value: publicEnv.PUBLIC_AUTH0_AUDIENCE }
		);
	}

	const missing = requiredEnvVars.filter(({ value }) => !value).map(({ name }) => name);

	if (missing.length > 0) {
		error(500, `Required environment variables are not defined: ${missing.join(', ')}`);
	}

	if (PUBLIC_AUTH_BACKEND !== SharedKey && PUBLIC_AUTH_BACKEND !== Auth0) {
		error(500, 'Invalid auth backend');
	}
};
