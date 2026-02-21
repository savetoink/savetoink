import { derived } from 'svelte/store';
import { page } from '$app/stores';

export const isLoggedIn = derived(page, ($page) => !!$page.data?.jwt);
