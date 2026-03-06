<script lang="ts">
	import { onMount } from 'svelte';
	import Account from './Account.svelte';
	import Post from './Post.svelte';
	import { Footer } from '@savetoink/shared/components';
	import { getAPIKey, getUserProfile } from '../../lib/storage';
	import '@savetoink/shared/css';
	import type { UserProfile } from '@savetoink/shared';

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
	<Footer />
</main>

<style>
	.container {
		min-width: 20rem;
		padding: 1.5rem;
	}
</style>
