export const STORAGE_KEY = "local:shared_api_key";
export const USER_PROFILE_KEY = "local:user_profile";

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
    const value = await storage.getItem<string>(STORAGE_KEY);
    return value ?? null;
  } catch (error) {
    console.error("failed to get API key from storage:", error);
    return null;
  }
}

export async function saveAPIKey(key: string): Promise<void> {
  try {
    await storage.setItem(STORAGE_KEY, key);
  } catch (error) {
    console.error("failed to save API key to storage:", error);
    throw error;
  }
}

export async function getUserProfile(): Promise<UserProfile | null> {
  try {
    const value = await storage.getItem<UserProfile>(USER_PROFILE_KEY);
    return value ?? null;
  } catch (error) {
    console.error("failed to get user profile from storage:", error);
    return null;
  }
}

export async function saveUserProfile(profile: UserProfile): Promise<void> {
  try {
    await storage.setItem(USER_PROFILE_KEY, profile);
  } catch (error) {
    console.error("failed to save user profile to storage:", error);
    throw error;
  }
}

export async function clearAPIKey(): Promise<void> {
  try {
    await storage.removeItem(STORAGE_KEY);
  } catch (error) {
    console.error("failed to remove API key from storage:", error);
    throw error;
  }
}

export async function clearUserProfile(): Promise<void> {
  try {
    await storage.removeItem(USER_PROFILE_KEY);
  } catch (error) {
    console.error("failed to remove user profile from storage:", error);
    throw error;
  }
}
