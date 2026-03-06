interface ImportMetaEnv {
	readonly PUBLIC_API_URL: string;
	readonly PUBLIC_APP_URL: string;
	readonly PUBLIC_AUTH_BACKEND: 'sharedKey' | 'auth0';
	readonly PUBLIC_AUTH0_CLIENT_ID: string;
	readonly PUBLIC_AUTH0_DOMAIN: string;
	readonly PUBLIC_AUTH0_AUDIENCE: string;
}

declare const __APP_VERSION__: string;
declare const __BUILD_DATE__: string;
declare const __GIT_HASH__: string;
