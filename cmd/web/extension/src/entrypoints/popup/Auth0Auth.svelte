<script lang="ts">
	import { getProfile, exchangeCodeForToken } from '../../lib/api';
	import type { components } from '@savetoink/shared';

	export let apiKey = '';
	export let profile: components['schemas']['UserProfile'] | null = null;
	export let onSave: (detail: {
		apiKey: string;
		profile?: components['schemas']['UserProfile'];
	}) => void | Promise<void>;
	export let onLogout: () => void | Promise<void>;

	let loginStatus = '';
	let redirectUri = '';

	$: isLoggedIn = apiKey && profile;

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
	<button type="button" on:click={handleLogin}> Log In / Sign Up </button>
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
	<button type="button" on:click={handleLogout}>Logout</button>
{/if}
