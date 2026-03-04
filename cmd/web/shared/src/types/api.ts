import type { UserProfile } from "./models";

export type AuthBackendType = "auth0" | "sharedKey";

export interface ProfileResponse {
  ok: boolean;
  status: number;
  profile?: UserProfile;
}

export interface CreateArticleResponse {
  id: string;
  title: string;
  url: string;
}

export interface ExchangeCodeResponse {
  access_token: string;
  expires_in?: number;
}
