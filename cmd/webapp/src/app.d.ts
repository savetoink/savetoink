// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		interface Locals {
			jwt?: string;
		}

		interface PageData {
			form?: {
				error?: string;
			};
		}

		interface Error {
			message: string;
		}
	}
}

export {};
