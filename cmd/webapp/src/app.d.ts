// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
import type { ApiClient } from '$lib/server/apiClient';

declare global {
	namespace App {
		interface Locals {
			apiClient: ApiClient | null;
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
