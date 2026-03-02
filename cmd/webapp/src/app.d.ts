declare global {
	namespace App {
		interface Locals {
			jwt?: string;
			isLoggedIn: boolean;
		}
		interface PageData {
			device_email?: string;
			auto_send?: boolean;
		}
	}
}

export {};
