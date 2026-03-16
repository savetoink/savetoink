<script lang="ts">
	import {
		PUBLIC_AUTH0_CLIENT_ID,
		PUBLIC_AUTH0_DOMAIN,
		PUBLIC_AUTH0_AUDIENCE,
		PUBLIC_APP_URL
	} from '$env/static/public';
	import { page } from '$app/state';
	import type { PageData } from './$types';

	const origin = PUBLIC_APP_URL || page.url.origin;
	const authUrl =
		`https://${PUBLIC_AUTH0_DOMAIN}/authorize?` +
		`response_type=code&` +
		`prompt=login&` +
		`client_id=${PUBLIC_AUTH0_CLIENT_ID}&` +
		`redirect_uri=${encodeURIComponent(`${origin}/auth/callback`)}&` +
		`audience=${PUBLIC_AUTH0_AUDIENCE}&` +
		`scope=openid profile email`;
	const login = () => (window.location.href = authUrl);

	let { data }: { data: PageData } = $props();
</script>

{#if data?.isLoggedIn}
	<p>Email: <code>{data.user?.email}</code></p>
	<form method="POST" action="?/clean">
		<button type="submit">Logout</button>
	</form>
{:else}
	<button onclick={login}>Log In / Sign Up</button>
{/if}
