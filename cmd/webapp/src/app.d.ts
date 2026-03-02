declare global {
	namespace App {
		interface Locals {
			jwt?: string;
			isLoggedIn: boolean;
			user?: {
				account: string;
				email: string;
				deviceEmail?: string;
				autoSend?: boolean;
			};
		}
	}
}

export {};
