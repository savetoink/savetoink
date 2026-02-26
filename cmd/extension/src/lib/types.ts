export interface AuthTokens {
  accessToken: string;
  refreshToken?: string;
  expiresAt: number;
  idToken?: string;
}

export interface UserProfile {
  sub: string;
  email: string;
  name?: string;
  picture?: string;
}

export interface AuthState {
  isAuthenticated: boolean;
  user: {
    id: string;
    email: string;
    name?: string;
    picture?: string;
  } | null;
}

export interface SendArticleResponse {
  id: string;
  title: string;
  url: string;
  message: string;
  deliveryStatus?: string;
}
