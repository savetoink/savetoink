import type { components } from '@savetoink/shared';

declare global {
	namespace App {
		interface Locals {
			auth?: string;
			isLoggedIn: boolean;
			user?: components['schemas']['UserProfile'];
		}
	}
}

export {};
