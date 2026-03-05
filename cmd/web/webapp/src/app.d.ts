import type { UserProfile } from '@savetoink/shared';

declare global {
	namespace App {
		interface Locals {
			auth?: string;
			isLoggedIn: boolean;
			user?: UserProfile;
		}
	}
}

export {};
