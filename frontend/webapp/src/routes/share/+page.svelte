<script lang="ts">
	import { enhance } from '$app/forms';
	import type { PageData } from './$types';

	export let data: PageData;
	const { incomingUrl } = data;
</script>

{#if !incomingUrl}
	<div class="card error">No URL received.</div>
{:else}
	<form method="POST" action="?/save" use:enhance>
		<input type="hidden" name="url" value={incomingUrl} />

		<div class="card">
			<p class="label">Save to your library?</p>
			<p class="url">{incomingUrl}</p>
			<button type="submit" class="btn-primary">Save</button>
		</div>
	</form>
{/if}
