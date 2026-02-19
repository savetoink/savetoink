<script lang="ts">
	import type { ActionData } from './$types';
	let { form }: { form: ActionData } = $props();
</script>

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
