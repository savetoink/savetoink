interface ImportMetaEnv {
  readonly VITE_SAVETOINK_API_URL: string;
  readonly VITE_SAVETOINK_AUTH_BACKEND: "shared_api_key" | "auth0";
  readonly VITE_SAVETOINK_AUTH0_CLIENT_ID: string;
  readonly VITE_SAVETOINK_AUTH0_DOMAIN: string;
  readonly VITE_SAVETOINK_AUTH0_AUDIENCE: string;
}
