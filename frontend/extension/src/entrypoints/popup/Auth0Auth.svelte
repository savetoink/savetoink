<script lang="ts">
	import { getProfile, exchangeCodeForToken } from '../../lib/api';
	import type { UserProfile } from '@savetoink/shared';

	let {
		apiKey,
		profile,
		onSave,
		onLogout
	}: {
		apiKey: string;
		profile: UserProfile | null;
		onSave: (detail: { apiKey: string; profile?: UserProfile }) => void | Promise<void>;
		onLogout: () => void | Promise<void>;
	} = $props();

	let isLoggedIn = $derived(!!apiKey && !!profile);
	let loginStatus = $state('');
	let redirectUri = '';

	async function handleLogin() {
		try {
			loginStatus = 'redirecting';

			redirectUri = browser.identity.getRedirectURL();

			const authUrl =
				`https://${import.meta.env.PUBLIC_AUTH0_DOMAIN}/authorize?` +
				`response_type=code&` +
				`prompt=login&` +
				`client_id=${import.meta.env.PUBLIC_AUTH0_CLIENT_ID}&` +
				`redirect_uri=${encodeURIComponent(redirectUri)}&` +
				`audience=${import.meta.env.PUBLIC_AUTH0_AUDIENCE}&` +
				`scope=openid profile email`;

			loginStatus = 'authenticating';

			const responseUrl = await browser.identity.launchWebAuthFlow({
				url: authUrl,
				interactive: true
			});

			if (!responseUrl) {
				throw new Error('authentication failed');
			}

			const url = new URL(responseUrl);
			const code = url.searchParams.get('code');

			if (!code) {
				throw new Error('no authorization code in redirect');
			}

			loginStatus = 'exchanging';

			const tokenResponse = await exchangeCodeForToken(code, redirectUri);

			loginStatus = 'fetching';

			const userProfile = await getProfile(tokenResponse.access_token);

			await onSave({
				apiKey: tokenResponse.access_token,
				profile: userProfile
			});

			loginStatus = 'saved';
			setTimeout(() => {
				loginStatus = '';
			}, 2000);
		} catch (error) {
			loginStatus = 'error';
			console.error('auth0 login failed:', error);
		}
	}

	async function handleLogout() {
		await onLogout();
	}
</script>

{#if !isLoggedIn}
	<button type="button" onclick={handleLogin}> Log In / Sign Up </button>
	{#if loginStatus === 'redirecting'}
		<p>redirecting to Auth0...</p>
	{:else if loginStatus === 'authenticating'}
		<p>authenticating with Auth0...</p>
	{:else if loginStatus === 'exchanging'}
		<p>exchanging authorization code for token...</p>
	{:else if loginStatus === 'fetching'}
		<p>fetching your profile...</p>
	{:else if loginStatus === 'saved'}
		<p>logged in successfully</p>
	{:else if loginStatus === 'error'}
		<p aria-invalid="true" class="error">authentication failed</p>
	{/if}
{:else}
	<button type="button" onclick={handleLogout}>Logout</button>
{/if}
