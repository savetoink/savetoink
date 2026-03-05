<script lang="ts">
	import { onMount } from 'svelte';
	import Account from './Account.svelte';
	import Post from './Post.svelte';
	import { getAPIKey, getUserProfile } from '../../lib/storage';
	import type { UserProfile } from '@savetoink/shared';
	import '@savetoink/shared/css';

	if (import.meta.env.DEV) {
		import('@savetoink/shared/css-dev');
	}

	let apiKey = $state('');
	let profile: UserProfile | null = $state(null);

	onMount(async () => {
		const [savedKey, savedProfile] = await Promise.all([getAPIKey(), getUserProfile()]);
		if (savedKey) {
			apiKey = savedKey;
		}
		profile = savedProfile;
	});
</script>

<main class="container">
	<section>
		<Post {profile} />
	</section>
	<section>
		<Account bind:profile bind:apiKey />
	</section>
</main>

<style>
	.container {
		min-width: 20rem;
		padding: 1.5rem;
	}
</style>
