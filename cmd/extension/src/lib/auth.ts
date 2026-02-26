import { storage } from '#imports';
import { createAuth0Client, Auth0Client } from '@auth0/auth0-spa-js';
import type { AuthTokens, UserProfile, AuthState } from './types';

const authTokens = storage.defineItem<AuthTokens | null>('local:authTokens', {
  fallback: null,
});

const userProfile = storage.defineItem<UserProfile | null>('local:userProfile', {
  fallback: null,
});

const authState = storage.defineItem<AuthState>('local:authState', {
  fallback: {
    isAuthenticated: false,
    user: null,
  },
});

let auth0Client: Auth0Client | null = null;

async function getAuth0Client(): Promise<Auth0Client> {
  if (!auth0Client) {
    auth0Client = await createAuth0Client({
      domain: import.meta.env.VITE_AUTH0_DOMAIN,
      clientId: import.meta.env.VITE_AUTH0_CLIENT_ID,
      authorizationParams: {
        redirect_uri: browser.identity.getRedirectURL(),
        scope: 'openid profile email offline_access',
        audience: import.meta.env.VITE_AUTH0_AUDIENCE,
      },
      cacheLocation: 'localstorage',
    });
  }
  return auth0Client;
}

export async function login(): Promise<void> {
  const client = await getAuth0Client();
  await client.loginWithRedirect();
}

export async function handleRedirectCallback(): Promise<void> {
  const client = await getAuth0Client();
  await client.handleRedirectCallback();

  const accessToken = await client.getTokenSilently();
  const user = await client.getUser();

  const tokens: AuthTokens = {
    accessToken,
    expiresAt: Date.now() + 3600 * 1000,
  };

  const profile: UserProfile = {
    sub: user?.sub || '',
    email: user?.email || '',
    name: user?.name,
    picture: user?.picture,
  };

  const state: AuthState = {
    isAuthenticated: true,
    user: {
      id: profile.sub,
      email: profile.email,
      name: profile.name,
      picture: profile.picture,
    },
  };

  await Promise.all([
    authTokens.setValue(tokens),
    userProfile.setValue(profile),
    authState.setValue(state),
  ]);

  browser.runtime.sendMessage({ type: 'AUTH_STATE_CHANGED', state });
}

export async function logout(): Promise<void> {
  const client = await getAuth0Client();
  await client.logout({
    logoutParams: {
      returnTo: browser.identity.getRedirectURL(),
    },
  });

  await Promise.all([
    authTokens.setValue(null),
    userProfile.setValue(null),
    authState.setValue({
      isAuthenticated: false,
      user: null,
    }),
  ]);

  browser.runtime.sendMessage({ type: 'AUTH_STATE_CHANGED', state: { isAuthenticated: false, user: null } });
}

export async function getAccessToken(): Promise<string | null> {
  const tokens = await authTokens.getValue();
  if (!tokens) return null;

  const now = Date.now();
  if (now >= tokens.expiresAt - 5 * 60 * 1000) {
    await refreshTokens();
    const refreshed = await authTokens.getValue();
    return refreshed?.accessToken ?? null;
  }

  return tokens.accessToken;
}

export async function isAuthenticated(): Promise<boolean> {
  const tokens = await authTokens.getValue();
  if (!tokens) return false;
  return Date.now() < tokens.expiresAt;
}

export async function refreshTokens(): Promise<void> {
  const client = await getAuth0Client();
  try {
    const accessToken = await client.getTokenSilently({
      authorizationParams: {
        scope: 'openid profile email offline_access',
      },
    });

    const tokens: AuthTokens = {
      accessToken,
      expiresAt: Date.now() + 3600 * 1000,
    };

    await authTokens.setValue(tokens);
  } catch (error) {
    console.error('Failed to refresh tokens:', error);
    await logout();
  }
}

export async function getAuthState(): Promise<AuthState> {
  return authState.getValue();
}

export function watchAuthState(callback: (state: AuthState) => void): () => void {
  return authState.watch((newState) => {
    callback(newState);
  });
}

export function sendMessage(type: string, data?: any): void {
  browser.runtime.sendMessage({ type, ...data });
}
