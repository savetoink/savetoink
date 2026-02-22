<script lang="ts">
	import { checkLoggedIn } from '$lib/auth';
	import type { ActionData, PageData } from './$types';

	let { form, data }: { form: ActionData; data: PageData } = $props();
	const loggedIn = $derived(checkLoggedIn(data));
	const shortJwt = $derived(data.jwt?.slice(0, 8) + '*'.repeat(data.jwt?.length - 8));
</script>

<section>
	{#if !loggedIn}
		<hgroup>
			<p>Enter your API key to access the article management system</p>
		</hgroup>
	{/if}

	{#if form?.error}
		<p class="error" role="alert">{form.error}</p>
	{/if}

	{#if loggedIn}
		<form method="POST" action="?/clean">
			<fieldset>
				<label>
					API Key
					<input type="text" name="jwt" value={shortJwt} />
				</label>
			</fieldset>
			<button type="submit">Logout</button>
		</form>
	{:else}
		<form method="POST" action="?/save">
			<fieldset>
				<label>
					API Key
					<input
						type="password"
						name="jwt"
						required
						autocomplete="current-password"
						placeholder="Enter your API key"
					/>
				</label>
			</fieldset>
			<button type="submit">Login</button>
		</form>
	{/if}
</section>
