import type { User } from '$lib/types';

declare global {
	namespace App {
		interface Locals {
			auth?: string;
			isLoggedIn: boolean;
			user?: User;
		}
	}
}

export {};
