import type { UserProfile } from '@savetoink/shared';

export const STORAGE_KEY = 'local:shared_api_key';
export const USER_PROFILE_KEY = 'local:user_profile';

function storageToUserProfile(profile: Partial<UserProfile> | null): UserProfile | null {
	if (!profile) return null;
	return {
		account: profile.account ?? '',
		email: profile.email ?? '',
		device_email: profile.device_email ?? '',
		auto_send: profile.auto_send ?? false
	};
}

function userProfileToStorage(profile: UserProfile): Partial<UserProfile> {
	return {
		account: profile.account,
		email: profile.email,
		device_email: profile.device_email,
		auto_send: profile.auto_send
	};
}

export async function getAPIKey(): Promise<string | null> {
	try {
		const value = await storage.getItem<string>(STORAGE_KEY);
		return value ?? null;
	} catch (error) {
		console.error('failed to get API key from storage:', error);
		return null;
	}
}

export async function saveAPIKey(key: string): Promise<void> {
	try {
		await storage.setItem(STORAGE_KEY, key);
	} catch (error) {
		console.error('failed to save API key to storage:', error);
		throw error;
	}
}

export async function getUserProfile(): Promise<UserProfile | null> {
	try {
		const value = await storage.getItem<Partial<UserProfile>>(USER_PROFILE_KEY);
		return storageToUserProfile(value);
	} catch (error) {
		console.error('failed to get user profile from storage:', error);
		return null;
	}
}

export async function saveUserProfile(profile: UserProfile): Promise<void> {
	try {
		await storage.setItem(USER_PROFILE_KEY, userProfileToStorage(profile));
	} catch (error) {
		console.error('failed to save user profile to storage:', error);
		throw error;
	}
}

export async function clearAPIKey(): Promise<void> {
	try {
		await storage.removeItem(STORAGE_KEY);
	} catch (error) {
		console.error('failed to remove API key from storage:', error);
		throw error;
	}
}

export async function clearUserProfile(): Promise<void> {
	try {
		await storage.removeItem(USER_PROFILE_KEY);
	} catch (error) {
		console.error('failed to remove user profile from storage:', error);
		throw error;
	}
}
