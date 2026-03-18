<script lang="ts">
	import { env as publicEnv } from '$env/dynamic/public';
	import { page } from '$app/state';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const login = () => {
		const origin = data?.publicAppUrl || page.url.origin;
		const authUrl =
			`https://${publicEnv.PUBLIC_AUTH0_DOMAIN}/authorize?` +
			`response_type=code&` +
			`prompt=login&` +
			`client_id=${publicEnv.PUBLIC_AUTH0_CLIENT_ID}&` +
			`redirect_uri=${encodeURIComponent(`${origin}/auth/callback`)}&` +
			`audience=${publicEnv.PUBLIC_AUTH0_AUDIENCE}&` +
			`scope=openid profile email`;
		window.location.href = authUrl;
	};
</script>

{#if data?.isLoggedIn}
	<p>Email: <code>{data.user?.email}</code></p>
	<form method="POST" action="?/clean">
		<button type="submit">Logout</button>
	</form>
{:else}
	<button onclick={login}>Log In / Sign Up</button>
{/if}
