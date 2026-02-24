<script lang="ts">
	import type { ActionData, PageData } from './$types';

	let { form, data }: { form: ActionData; data: PageData } = $props();
	const loggedIn = $derived(data.isLoggedIn);
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
					<input type="text" name="jwt" value={data.jwt} />
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
