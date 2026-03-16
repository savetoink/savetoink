import { handle as rootHandle } from '$lib/server/hooks';
import type { Handle } from '@sveltejs/kit';

export const handle: Handle = rootHandle;
