/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly PUBLIC_APP_URL: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
