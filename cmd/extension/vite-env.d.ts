interface ImportMetaEnv {
  readonly PUBLIC_API_URL: string;
  readonly PUBLIC_AUTH_BACKEND: "sharedKey" | "auth0";
  readonly PUBLIC_AUTH0_CLIENT_ID: string;
  readonly PUBLIC_AUTH0_DOMAIN: string;
  readonly PUBLIC_AUTH0_AUDIENCE: string;
}
