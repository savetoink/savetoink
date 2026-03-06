import type { UserProfile } from '@savetoink/shared';

export const STORAGE_KEY = 'local:shared_api_key';
export const USER_PROFILE_KEY = 'local:user_profile';

// why don't we use the shared types?
interface ApiUserProfile {
	account: string;
	email: string;
	device_email?: string;
	auto_send?: boolean;
}

function apiToSharedProfile(apiProfile: ApiUserProfile): UserProfile {
	return {
		account: apiProfile.account,
		email: apiProfile.email,
		device_email: apiProfile.device_email ?? '',
		auto_send: apiProfile.auto_send ?? false
	};
}

function sharedToApiProfile(sharedProfile: UserProfile): ApiUserProfile {
	return {
		account: sharedProfile.account,
		email: sharedProfile.email,
		device_email: sharedProfile.device_email,
		auto_send: sharedProfile.auto_send
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
		const value = await storage.getItem<ApiUserProfile>(USER_PROFILE_KEY);
		return value ? apiToSharedProfile(value) : null;
	} catch (error) {
		console.error('failed to get user profile from storage:', error);
		return null;
	}
}

export async function saveUserProfile(profile: UserProfile): Promise<void> {
	try {
		await storage.setItem(USER_PROFILE_KEY, sharedToApiProfile(profile));
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
