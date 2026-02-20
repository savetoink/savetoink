<script lang="ts">
	import type { ActionData } from './$types';
	let { form }: { form: ActionData } = $props();
</script>

<section>
	<hgroup>
		<h1>Login</h1>
		<p>Enter your API key to access the article management system</p>
	</hgroup>

	{#if form?.error}
		<p class="error" role="alert">{form.error}</p>
	{/if}

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
		<button type="submit">Save</button>
	</form>

	<form method="POST" action="?/clean">
		<button type="submit">Clean</button>
	</form>
</section>
