<script lang="ts">
	import {
		PUBLIC_AUTH_BACKEND,
		PUBLIC_AUTH0_CLIENT_ID,
		PUBLIC_AUTH0_DOMAIN,
		PUBLIC_AUTH0_AUDIENCE
	} from '$env/static/public';
	import { page } from '$app/state';
	import type { PageProps } from './$types';

	const origin = page.url.origin;
	const authUrl =
		`https://${PUBLIC_AUTH0_DOMAIN}/authorize?` +
		`response_type=code&` +
		`prompt=login&` +
		`client_id=${PUBLIC_AUTH0_CLIENT_ID}&` +
		`redirect_uri=${encodeURIComponent(`${origin}/auth/callback`)}&` +
		`audience=${PUBLIC_AUTH0_AUDIENCE}&` +
		`scope=openid profile email`;
	const login = () => (window.location.href = authUrl);
	const logout = async () => {
		await fetch('?/clean', { method: 'POST', body: new FormData() });
		window.location.href =
			`https://${PUBLIC_AUTH0_DOMAIN}/logout?` +
			`client_id=${PUBLIC_AUTH0_CLIENT_ID}&` +
			`returnTo=${encodeURIComponent(origin)}`;
	};
	let { form }: PageProps = $props();
</script>

{#if PUBLIC_AUTH_BACKEND === 'auth0'}
	<button onclick={login}>Login with Auth0</button>
	<button onclick={logout}>Logout</button>
{:else}
	<h1>Login</h1>
	<p>Enter your API key to access the article management system.</p>

	{#if form?.error}
		<p class="error">{form.error}</p>
	{/if}

	<form method="POST" action="?/save">
		<label>
			API Key
			<input type="password" name="jwt" required autocomplete="current-password" />
		</label>
		<button type="submit">Save</button>
	</form>

	<form method="POST" action="?/clean">
		<button type="submit">Clean</button>
	</form>
{/if}
