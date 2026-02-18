<script lang="ts">
	import {
		PUBLIC_AUTH0_CLIENT_ID,
		PUBLIC_AUTH0_DOMAIN,
		PUBLIC_AUTH0_AUDIENCE
	} from '$env/static/public';
	import { page } from '$app/state';

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
</script>

<button onclick={login}>Login with Auth0</button>
<button onclick={logout}>Logout</button>
