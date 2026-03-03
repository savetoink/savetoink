export const STORAGE_KEY = "shared_api_key";
export const AUTH_BACKEND_KEY = "auth_backend";
export const USER_PROFILE_KEY = "user_profile";

export type AuthBackendType = "shared_api_key" | "auth0";

export interface UserProfile {
  account: string;
  email: string;
  device_email: string;
  auto_send: boolean;
}
export const SharedKeyBackend: AuthBackendType = "shared_api_key";
export const Auth0Backend: AuthBackendType = "auth0";

export async function getAPIKey(): Promise<string | null> {
  try {
    const result = await browser.storage.sync.get(STORAGE_KEY);
    const value = result[STORAGE_KEY];
    return typeof value === "string" ? value : null;
  } catch (error) {
    console.error("failed to get API key from storage:", error);
    return null;
  }
}

export async function saveAPIKey(key: string): Promise<void> {
  try {
    await browser.storage.sync.set({ [STORAGE_KEY]: key });
  } catch (error) {
    console.error("failed to save API key to storage:", error);
    throw error;
  }
}

export async function getAuthBackend(): Promise<AuthBackendType> {
  try {
    const result = await browser.storage.sync.get(AUTH_BACKEND_KEY);
    const value = result[AUTH_BACKEND_KEY];
    return value === Auth0Backend ? Auth0Backend : SharedKeyBackend;
  } catch (error) {
    console.error("failed to get auth backend from storage:", error);
    return SharedKeyBackend;
  }
}

export async function saveAuthBackend(type: AuthBackendType): Promise<void> {
  try {
    await browser.storage.sync.set({ [AUTH_BACKEND_KEY]: type });
  } catch (error) {
    console.error("failed to save auth backend to storage:", error);
    throw error;
  }
}

export async function getUserProfile(): Promise<UserProfile | null> {
  try {
    const result = await browser.storage.sync.get(USER_PROFILE_KEY);
    const value = result[USER_PROFILE_KEY];
    if (value && typeof value === "object") {
      return value as UserProfile;
    }
    return null;
  } catch (error) {
    console.error("failed to get user profile from storage:", error);
    return null;
  }
}

export async function saveUserProfile(profile: UserProfile): Promise<void> {
  try {
    await browser.storage.sync.set({ [USER_PROFILE_KEY]: profile });
  } catch (error) {
    console.error("failed to save user profile to storage:", error);
    throw error;
  }
}

export async function clearAPIKey(): Promise<void> {
  try {
    await browser.storage.sync.remove(STORAGE_KEY);
  } catch (error) {
    console.error("failed to remove API key from storage:", error);
    throw error;
  }
}

export async function clearUserProfile(): Promise<void> {
  try {
    await browser.storage.sync.remove(USER_PROFILE_KEY);
  } catch (error) {
    console.error("failed to remove user profile from storage:", error);
    throw error;
  }
}
