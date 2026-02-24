declare global {
	namespace App {
		interface Locals {
			jwt?: string;
			isLoggedIn: boolean;
		}
	}
}

export {};
